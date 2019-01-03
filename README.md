# hermes

Yet Another lightweight, distributed, efficient and fast In-Memory Key-Value Store alternative that uses [Bloom Filter](https://en.wikipedia.org/wiki/Bloom_filter), [HAMT](https://en.wikipedia.org/wiki/Hash_array_mapped_trie) map, [LRFU](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.57.2214&rep=rep1&type=pdf) Cache Removal Policy, Load-Balancing thru [Balls-into-bins](https://arxiv.org/pdf/1608.01350.pdf) stochastic process, peer observation propagation and discovery via [Smudge](https://github.com/clockworksoul/smudge) and a faster consistent non-cryptographic hashing algorithm called [Fast Positive Hash (*Позитивный Хэш*)](https://github.com/leo-yuriev/t1ha).


This was a originally sub-component of a pet stuff I do. I thought this is helpful enough and can stand on its own to help alleviate a slow app due to an abusive disk I/O operations (database, file).


Visit the /hermes/server implementation Readme or read about it [here](https://myth-of-sissyphus.blogspot.com/2018/10/hermes-yet-another-one-key-value-store.html)
