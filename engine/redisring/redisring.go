package redisring

import (
	"log"
	"time"

	"github.com/pkg/errors"
	redis "gopkg.in/redis.v4"
)

// Engine uses redis.v4 as the back end
type Engine struct {
	prefix string
	ring   *redis.Ring

	shouldLogErrors bool
	cleanupTimeout  time.Duration
}

const expirePrefix = "expire:"
const lockPrefix = "lock:"

// NewRedisRingEngine creates a new redis ring for use as a store
func NewRedisRingEngine(
	prefix string,
	ringOpts *redis.RingOptions,
	cleanupTimeout time.Duration,
	shouldLogErrors bool,
) (*Engine, error) {
	if ringOpts == nil {
		return nil, errors.New("nil ringOpts passed to NewRedisRingEngine")
	}

	if len(ringOpts.Addrs) == 0 {
		return nil, errors.New("redisring options must have 1 or more addresses")
	}

	return &Engine{
		prefix:          prefix + ":",
		ring:            redis.NewRing(ringOpts),
		cleanupTimeout:  cleanupTimeout,
		shouldLogErrors: shouldLogErrors,
	}, nil
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	err := e.hasRing("Exists")
	if err != nil {
		return false
	}

	cmd := e.ring.Exists(e.prefix + key)
	result, err := cmd.Result()
	if err != nil {
		if e.shouldLogErrors {
			log.Println(err.Error())
		}
		return false
	}
	return result
}

// Get retrieves data from teh store based on the key if it exists,
// returns an error if the key does not exist or the redis connection fails
func (e *Engine) Get(key string) ([]byte, error) {
	err := e.hasRing("Get")
	if err != nil {
		return nil, err
	}

	cmd := e.ring.Get(e.prefix + key)
	result, err := cmd.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "attempting to get") // TODO add key
	}

	return result, nil
}

// Put stores data against a key, else it returns an error
// SETEX doesn't exist within this lib, it's advised to use Set for similar behavior
// https://github.com/go-redis/redis/blob/dc9d5006b3c319de24b2fa4de242e442553fcce2/commands.go#L726
func (e *Engine) Put(key string, data []byte, expires time.Time) error {
	err := e.hasRing("Put")
	if err != nil {
		return err
	}

	dataCmd := e.ring.Set(e.prefix+key, data, e.cleanupTimeout)
	err = dataCmd.Err()
	if err != nil {
		return errors.Wrap(err, "attempting to set data") // TODO add key
	}

	expireCmd := e.ring.Set(e.prefix+expirePrefix+key, data, e.cleanupTimeout)
	err = expireCmd.Err()
	if err != nil {
		return errors.Wrap(err, "attempting to set expire key") // TODO add key
	}

	return nil
}

// IsExpired checks to see if the given key has expired
func (e *Engine) IsExpired(key string) bool {
	err := e.hasRing("IsExpired")
	if err != nil {
		return false
	}

	if e.Exists(expirePrefix + key) {
		k := e.prefix + expirePrefix + key
		cmd := e.ring.Get(k)
		result, err := cmd.Int64()
		if err != nil {
			if e.shouldLogErrors {
				log.Println("error checking expired for key: " + k)
			}
			return false
		}

		if time.Now().Unix() > result {
			return true
		}
	}

	return false
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	err := e.hasRing("IsLocked")
	if err != nil {
		return false
	}

	return e.Exists(lockPrefix + key)
}

// Lock sets a lock against a given key
// SETEX doesn't exist within this lib, it's advised to use Set for similar behavior
// https://github.com/go-redis/redis/blob/dc9d5006b3c319de24b2fa4de242e442553fcce2/commands.go#L726
func (e *Engine) Lock(key string) error {
	err := e.hasRing("Lock")
	if err != nil {
		return err
	}

	k := e.getLockKey(lockPrefix, key)
	cmd := e.ring.Set(k, []byte("1"), e.cleanupTimeout)

	err = cmd.Err()
	if err != nil {
		return errors.Wrap(err, "attempting to lock") // TODO add key
	}

	return nil
}

// Unlock removes the lock from a given key
func (e *Engine) Unlock(key string) error {
	err := e.hasRing("Unlock")
	if err != nil {
		return err
	}

	k := e.getLockKey(lockPrefix, key)
	cmd := e.ring.Del(k)

	err = cmd.Err()
	if err != nil {
		return errors.Wrap(err, "attempting to unlock") // TODO add key
	}

	return nil
}

// Expire marks the key as expired and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	err := e.hasRing("Expire")
	if err != nil {
		return err
	}

	// delete all relevant keys
	cmd := e.ring.Del(
		e.prefix+key,
		e.prefix+expirePrefix+key,
		e.prefix+lockPrefix+key,
	)

	err = cmd.Err()
	if err != nil {
		return errors.Wrap(err, "attempted to expire") // TODO add key
	}

	return nil
}

// helper function that checks to see if a valid ring exists on the engine
func (e *Engine) hasRing(method string) error {
	if e.ring != nil {
		return nil
	}

	err := errors.New(method + ": nil ring in redisring engine")
	if e.shouldLogErrors {
		log.Println(err.Error())
	}
	return err
}

// helper function for locking / unlocking keys
func (e *Engine) getLockKey(lockPrefix, key string) string {
	return e.prefix + lockPrefix + key
}
