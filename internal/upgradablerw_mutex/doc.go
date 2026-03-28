// Package upgradable_rw_mutex provides an [UpgradableRWMutex] that extends
// sync.RWMutex with the ability to atomically upgrade a read lock to a write
// lock without releasing it first.
//
// Multiple goroutines may hold concurrent read locks alongside a single
// upgradable-read lock. The holder of the upgradable-read lock can call
// [UpgradableRWMutex.UpgradeWLock] to promote to exclusive write access,
// blocking until all other readers have released.
//
// # Layer
//
// Infrastructure — concurrency primitive with no internal dependencies.
package upgradable_rw_mutex
