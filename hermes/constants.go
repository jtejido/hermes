package hermes

const (
	defaultBasePath      = "/_hermes/"
	defaultReplicas      = 10
	maxCuckooCount       = 500
	nullFp               = byte(0)
	bucketSize           = 4
	timestampSizeInBytes = 8                                                       // Number of bytes used for timestamp
	hashSizeInBytes      = 8                                                       // Number of bytes used for hash
	keySizeInBytes       = 2                                                       // Number of bytes used for size of entry key
	headersSizeInBytes   = timestampSizeInBytes + hashSizeInBytes + keySizeInBytes // Number of bytes used for all headers
)
