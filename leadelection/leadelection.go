package leadelection

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/utils/wait"
)

// ErrNotFound means item not found in nacos
var ErrNotFound = errors.New("item not found")

const (
	jitterFactor = 1.2
)

// LeaderElectionRecord is a record for leader election operation
type LeaderElectionRecord struct {
	// HolderIdentity is the ID that owns the lease. If empty, no one owns this lease and
	// all callers may acquire. Versions of this library prior to Kubernetes 1.14 will not
	// attempt to acquire leases with empty identities and will wait for the full lease
	// interval to expire before attempting to reacquire. This value is set to empty when
	// a client voluntarily steps down.
	HolderIdentity       string `json:"holderIdentity"`
	LeaseDurationSeconds int    `json:"leaseDurationSeconds"`
	AcquireTime          int64  `json:"acquireTime"`
	RenewTime            int64  `json:"renewTime"`
	LeaderTransitions    int    `json:"leaderTransitions"`
}

// ILock is a holder for LeaderElectionRecord
type ILock interface {
	// Get returns the LeaderElectionRecord
	Get(ctx context.Context) (*LeaderElectionRecord, []byte, error)

	// Create attempts to create a LeaderElectionRecord
	Create(ctx context.Context, ler LeaderElectionRecord) error

	// Update will update and existing LeaderElectionRecord
	Update(ctx context.Context, ler LeaderElectionRecord) error

	// RecordEvent is used to record events
	RecordEvent(string)

	// Identity will return the locks Identity
	Identity() string

	// Describe is used to convert details on current resource lock
	// into a string
	Describe() string
}

// NewLeaderElector creates a LeaderElector from a LeaderElectionConfig
func NewLeaderElector(lec LeaderElectionConfig, logger logging.ILogger) (*LeaderElector, error) {
	if lec.LeaseDuration <= lec.RenewDeadline {
		return nil, fmt.Errorf("leaseDuration must be greater than renewDeadline")
	}
	if lec.RenewDeadline <= time.Duration(jitterFactor*float64(lec.RetryPeriod)) {
		return nil, fmt.Errorf("renewDeadline must be greater than retryPeriod*JitterFactor")
	}
	if lec.LeaseDuration < 1 {
		return nil, fmt.Errorf("leaseDuration must be greater than zero")
	}
	if lec.RenewDeadline < 1 {
		return nil, fmt.Errorf("renewDeadline must be greater than zero")
	}
	if lec.RetryPeriod < 1 {
		return nil, fmt.Errorf("retryPeriod must be greater than zero")
	}
	if lec.Callbacks.OnStartedLeading == nil {
		return nil, fmt.Errorf("the OnStartedLeading callback must not be nil")
	}
	if lec.Callbacks.OnStoppedLeading == nil {
		return nil, fmt.Errorf("the OnStoppedLeading callback must not be nil")
	}
	if lec.Lock == nil {
		return nil, fmt.Errorf("the Lock must not be nil")
	}
	le := LeaderElector{
		config: lec,
		logger: logger,
	}
	return &le, nil
}

// LeaderElectionConfig is config for LeaderElector
type LeaderElectionConfig struct {
	// Lock is the resource that will be used for locking
	Lock ILock

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	//
	// A client needs to wait a full LeaseDuration without observing a change to
	// the record before it can attempt to take over. When all clients are
	// shutdown and a new set of clients are started with different names against
	// the same leader record, they must wait the full LeaseDuration before
	// attempting to acquire the lease,Thus LeaseDuration should be as short as
	// possible (within your tolerance for clock skew rate) to avoid a possible
	// long waits in the scenario.
	//
	// Core clients default this value to 15 seconds.
	LeaseDuration time.Duration
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	//
	// Core clients default this value to 10 seconds.
	RenewDeadline time.Duration
	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	//
	// Core clients default this value to 2 seconds.
	RetryPeriod time.Duration

	// Callbacks are callbacks that are triggered during certain lifecycle
	// events of the LeaderElector
	Callbacks LeaderCallbacks

	// ReleaseOnCancel should be set true if the lock should be released
	// when the run context is cancelled. If you set this to true, you must
	// ensure all code guarded by this lease has successfully completed
	// prior to cancelling the context, or you may have two processes
	// simultaneously acting on the critical path.
	ReleaseOnCancel bool

	// Name is the name of the resource lock for debugging
	Name string
}

// LeaderCallbacks are callbacks that are triggered during certain
// lifecycle events of the LeaderElector. These are invoked asynchronously.
//
// possible future callbacks:
//   - OnChallenge()
type LeaderCallbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading
	OnStartedLeading func(context.Context)
	// OnStoppedLeading is called when a LeaderElector client stops leading
	OnStoppedLeading func()
	// OnNewLeader is called when the client observes a leader that is
	// not the previously observed leader. This includes the first observed
	// leader when the client starts.
	OnNewLeader   func(identity string)
	OnStopRunning func()
}

// LeaderElector is a leader election client.
type LeaderElector struct {
	logger logging.ILogger
	config LeaderElectionConfig
	// internal bookkeeping
	observedRecord    LeaderElectionRecord
	observedRawRecord []byte
	observedTime      time.Time
	// used to implement OnNewLeader(), may lag slightly from the
	// value observedRecord.HolderIdentity if the transition has
	// not yet been reported.
	reportedLeader string

	// used to lock the observedRecord
	observedRecordLock sync.Mutex
}

// Run starts the leader election loop. Run will not return
// before leader election loop is stopped by ctx, or it has
// stopped holding the leader lease
func (le *LeaderElector) Run(ctx context.Context) {
	fmt.Println("=== START LEADER ELECTION LOOP ===")
	defer func() {
		le.config.Callbacks.OnStoppedLeading()
	}()

	if !le.acquire(ctx) {
		return // ctx signalled done
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go le.config.Callbacks.OnStartedLeading(ctx)
	le.renew(ctx)
	fmt.Println("=== STOP LEADER ELECTION LOOP ===")
	le.config.Callbacks.OnStopRunning()
}

// RunOrDie starts a client with the provided config or panics if the config
// fails to validate. RunOrDie blocks until leader election loop is
// stopped by ctx ,or it has stopped holding the leader lease
//
//goland:noinspection GoUnusedExportedFunction
func RunOrDie(ctx context.Context, lec LeaderElectionConfig, logger logging.ILogger) {
	le, err := NewLeaderElector(lec, logger)
	if err != nil {
		panic(err)
	}
	le.Run(ctx)
}

// GetLeader returns the identity of the last observed leader or returns the empty string if
// no leader has yet been observed.
// This function is for informational purposes. (e.g. monitoring, logs, etc.)
func (le *LeaderElector) GetLeader() string {
	return le.getObservedRecord().HolderIdentity
}

// IsLeader returns true if the last observed leader was this client else returns false.
func (le *LeaderElector) IsLeader() bool {
	return le.getObservedRecord().HolderIdentity == le.config.Lock.Identity()
}

// acquire loops calling tryAcquireOrRenew and returns true immediately when tryAcquireOrRenew succeeds.
// Returns false if ctx signals done.
func (le *LeaderElector) acquire(ctx context.Context) bool {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	succeeded := false
	desc := le.config.Lock.Describe()
	le.logger.Debugf("attempting to acquire leader lease %v...", desc)
	wait.JitterUntil(func() {
		succeeded = le.tryAcquireOrRenew(ctx)
		le.maybeReportTransition()
		if !succeeded {
			le.logger.Debugf("failed to acquire lease %v", desc)
			return
		}
		le.config.Lock.RecordEvent("became leader")
		le.logger.Debugf("successfully acquired lease %v", desc)
		cancel()
	}, le.config.RetryPeriod, jitterFactor, true, ctx.Done())
	return succeeded
}

// renew loops calling tryAcquireOrRenew and returns immediately when tryAcquireOrRenew fails or ctx signals done.
func (le *LeaderElector) renew(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wait.Until(func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, le.config.RenewDeadline)
		defer timeoutCancel()
		err := wait.PollImmediateUntil(le.config.RetryPeriod, func() (bool, error) {
			return le.tryAcquireOrRenew(timeoutCtx), nil
		}, timeoutCtx.Done())

		le.maybeReportTransition()
		desc := le.config.Lock.Describe()
		if err == nil {
			le.logger.Debugf("successfully renewed lease %v", desc)
			return
		}
		le.config.Lock.RecordEvent("stopped leading")
		le.logger.Debugf("failed to renew lease %v: %v", desc, err)
		cancel()
	}, le.config.RetryPeriod, ctx.Done())

	// if we hold the lease, give it up
	if le.config.ReleaseOnCancel {
		le.release()
	}
}

// release attempts to release the leader lease if we have acquired it.
func (le *LeaderElector) release() bool {
	if !le.IsLeader() {
		return true
	}
	now := time.Now().UTC()
	leaderElectionRecord := LeaderElectionRecord{
		LeaderTransitions:    le.observedRecord.LeaderTransitions,
		LeaseDurationSeconds: 1,
		RenewTime:            now.Unix(),
		AcquireTime:          now.Unix(),
	}
	if err := le.config.Lock.Update(context.TODO(), leaderElectionRecord); err != nil {
		le.logger.Errorf("Failed to release lock: %v", err)
		return false
	}

	le.setObservedRecord(&leaderElectionRecord)
	return true
}

// tryAcquireOrRenew tries to acquire a leader lease if it is not already acquired,
// else it tries to renew the lease if it has already been acquired. Returns true
// on success else returns false.
func (le *LeaderElector) tryAcquireOrRenew(ctx context.Context) bool {
	now := time.Now().UTC()
	leaderElectionRecord := LeaderElectionRecord{
		HolderIdentity:       le.config.Lock.Identity(),
		LeaseDurationSeconds: int(le.config.LeaseDuration / time.Second),
		RenewTime:            now.Unix(),
		AcquireTime:          now.Unix(),
	}

	// 1. obtain or create the ElectionRecord
	oldLeaderElectionRecord, oldLeaderElectionRawRecord, err := le.config.Lock.Get(ctx)
	if err != nil {

		if err != ErrNotFound {
			le.logger.Errorf("error retrieving resource lock %v: %v", le.config.Lock.Describe(), err)
			return false
		}

		if err = le.config.Lock.Create(ctx, leaderElectionRecord); err != nil {
			le.logger.Errorf("error initially creating leader election record: %v", err)
			return false
		}

		le.setObservedRecord(&leaderElectionRecord)

		return true
	}

	// 2. Record obtained, check the Identity & Time
	if !bytes.Equal(le.observedRawRecord, oldLeaderElectionRawRecord) {
		le.setObservedRecord(oldLeaderElectionRecord)

		le.observedRawRecord = oldLeaderElectionRawRecord
	}

	if len(oldLeaderElectionRecord.HolderIdentity) > 0 &&
		time.Unix(oldLeaderElectionRecord.RenewTime, 0).Add(le.config.LeaseDuration).After(now) &&
		!le.IsLeader() {
		le.logger.Debugf("lock is held by %v and has not yet expired", oldLeaderElectionRecord.HolderIdentity)
		return false
	}

	// 3. We're going to try to update. The leaderElectionRecord is set to it's default
	// here. Let's correct it before updating.
	if le.IsLeader() {
		leaderElectionRecord.AcquireTime = oldLeaderElectionRecord.AcquireTime
		leaderElectionRecord.LeaderTransitions = oldLeaderElectionRecord.LeaderTransitions
	} else {
		leaderElectionRecord.LeaderTransitions = oldLeaderElectionRecord.LeaderTransitions + 1
	}

	// update the lock itself
	if err = le.config.Lock.Update(ctx, leaderElectionRecord); err != nil {
		le.logger.Errorf("Failed to update lock: %v", err)
		return false
	}

	le.setObservedRecord(&leaderElectionRecord)
	return true
}

func (le *LeaderElector) maybeReportTransition() {
	if le.observedRecord.HolderIdentity == le.reportedLeader {
		return
	}
	le.reportedLeader = le.observedRecord.HolderIdentity
	if le.config.Callbacks.OnNewLeader != nil {
		go le.config.Callbacks.OnNewLeader(le.reportedLeader)
	}
}

// Check will determine if the current lease is expired by more than timeout.
func (le *LeaderElector) Check(maxTolerableExpiredLease time.Duration) error {
	if !le.IsLeader() {
		// Currently not concerned with the case that we are hot standby
		return nil
	}
	// If we are more than timeout seconds after the lease duration that is past the timeout
	// on the lease renew. Time to start reporting ourselves as unhealthy. We should have
	// died but conditions like deadlock can prevent this. (See #70819)
	if time.Since(le.observedTime) > le.config.LeaseDuration+maxTolerableExpiredLease {
		return fmt.Errorf("failed election to renew leadership on lease %s", le.config.Name)
	}

	return nil
}

// setObservedRecord will set a new observedRecord and update observedTime to the current time.
// Protect critical sections with lock.
func (le *LeaderElector) setObservedRecord(observedRecord *LeaderElectionRecord) {
	le.observedRecordLock.Lock()
	defer le.observedRecordLock.Unlock()

	le.observedRecord = *observedRecord
	le.observedTime = time.Now().UTC()
}

// getObservedRecord returns observersRecord.
// Protect critical sections with lock.
func (le *LeaderElector) getObservedRecord() LeaderElectionRecord {
	le.observedRecordLock.Lock()
	defer le.observedRecordLock.Unlock()

	return le.observedRecord
}
