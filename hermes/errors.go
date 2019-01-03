package hermes

import (
	"fmt"
)

var (
	shardsNotInitializedError = "Shards not initialized."
	policyNotInitializedError = "Policy not initialized."
	shardNotFoundForKeyError  = "Shard not found for key: %s."
	keyNotFoundInShardError   = "Item with key: '%s' not found at shard: %d"
	loaderError               = "Loader error: %v"
	collisionDetectedError    = "Collision detected: key: '%s' not found at shard: %d"
	filterFirstInstanceError  = "Not found in filter. First instance for key: '%s'"
)

// any message above, and corresponding arguments
func errorf(msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}
