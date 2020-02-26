# Cryptography

This package contains all the cryptography tools needed in Flow across all the protocol streams.  
Some tools offer cryptographic services that can be used in other contexts and projects while others are specific to Flow. 

## Hashing

A cryptographic hash function is needed in Flow to provide different services:

- It verifies message integrity of communications and prevents some non-malicious and malicious alterations.
- It is involved in signature schemes during signature generations and signature verifications. The signature and verification algorithms are applied over the short message digests instead of the original message. 
- It generates pseudo random data from inputs that might have dispersed entropy and might not be distributed uniformly.

A more generic (non-cryptographic) hash function can be used as a data identifier where the protection against forgery is not required. Data digests used as identifiers allow services such as storage space optimization or faster data lookups.  
A cryptographic hash function can be used in this context although it is more computationally expensive. 

### SHA3

[SHA3](https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.202.pdf) is used as the cryptographic hash function in Flow. 
[SHA3-384] is used to hash messages prior to applying the signature algorithm. The digest size is 384 bits which is suitable to the signature algorithm used. 
[SHA3-256] is used when signatures are not invloved. The digest size is 256 bits while the input message can be of arbitrary size. The algorithm provides a 128-bits collision resistance, along with a 256-bits security strength for preimage and 2nd preimage.

### Interface

The hash interface allows instantiating different hash algorithms and using them together in the same protocol flow. 

The hashing interface mainly supports the hash computation of some arbitrary data (a byte array). The digest is of type `Hash`. 

The interface also supports appending data to the hash state without computing the final hash. The hash computation can be finalized only when the complete data stream has been fed to the hash state. It is possible to reset the hash state to empty the previously added data. 


## The Signature Scheme

A signature scheme is needed in Flow for several purposes:
- The main purpose is to provide authentication to communications between nodes. 
- It provides a way to generate a distributed random number through [threshold signatures](#Threshold-Signature). 
- In a specific [use-case in Flow](#verification-of-ownership), it provides a verifiable proof of the ownership of some secret data without revealing the secret data itself. 

A signature scheme is defined by three functions:
- Key generation: Given a seed, the function generates a private key and a public key.
- Signature generation: Given a private key and a message, the function generates a signature.
- Signature verification: Given a signature and a public key, the function verifies if the signature is valid (i.e. it was generated using the private key associated with the given public key).

### BLS

BLS is a signature scheme named after Boneh, Lynn, and Shacha.  
The signature generated by BLS is unique (i.e. only one signature verifies per a message and a public key).   
It provides a non-interactive threshold signature, which is also unique, along with a simple aggregation signature. 

In BLS, a private key is a field element while a public key is an element of a cyclic group G1. A signature is an element of a second cyclic group G2.  
Signing is an exponentiation in G1 while a verification is an equility check of two pairings results. 

### BLS12-381

[BLS12-381](https://github.com/zkcrypto/pairing/tree/master/src/bls12_381) is an instance of the BLS curves family named after Baretto, Lynn, and Scott. It is a pairing-friendly curve and offers a bit security level close to 128. 

### Interface

The signature interface allows using multiple signature schemes.

A key generation function generates a key pair given a seed.  
The signature interface signs a message given a key pair (only the private key is used in BLS) and a hash algorithm instance. The signature is of a type `Signature`.  
The signature verification interface verifies the validity of a signature given a public key and the message used for signing. 

### Aggregation

Signature aggregation means generating a single signature output in one of the following scenarios:
- multiple signers signing the same single message
- multiple signers signing each a different message
- a single signer signing multiple messages

Although the aggregated signature is of a size of a single signature, the aggregated signature verification is different in each scenario and can be heavier than verifying a single signature. 

BLS provides a simple aggregation solution. The aggregation is non-interactive and it doesn't require knowledge of any private key or the message(s) being signed. It therefore can be performed by any party, whether they are signers or not. 

The aggregation should be used in Flow in different streams, it is still not decided what scenarios will be used in what part of the streams. The interface will be defined accordingly.
 

## Distributed Random Generation

Flow requires a source of randomness for several mechanisms needed in the protocol. 

At each Epoch, the generation process needs to output a sequence of pseudo-random numbers that are deterministic, unpredictable by any party until the random number itself is generated, agreed upon by all nodes, and verifiable by anyone against the commitment. 

### Distributed Key Generation

The DKG is run every time a new pseudo-random numbers sequence is needed. It is run by a subset of _n_ nodes and generates a uniformly distributed key pair using entropy from all the _n_ nodes.  
The private key is not known by any party while the public key is known by all _n_ parties.  
Only shares of the private key are known by the parties, each having a secret share. 

The protocol is interactive, and is robust if the number of malicious nodes does not exceed a threshold _t_. It is possible for _t+1_ parties to construct the secret private key by combining their _t+1_ secret shares. 

The DKG protocol to be implemented is the [Gennaro et al.](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.50.2737&rep=rep1&type=pdf) protocol.

### Threshold Signature

A threshold signature is the operation of signing a message by a group of _n_ parties, given the group has already generated an unkown private key and each party got a secret share of that private key.  
The signature can only be generated if _t+1_ parties or more have signed the message using their secret shares. The output signature is verifiable by the group public key and does not depend on the subgroup that signed the message. 

BLS provides a non-interactive [protocol to generate threshold signatures](https://www.iacr.org/archive/pkc2003/25670031/25670031.pdf). These signatures inherit the uniqueness property from the single BLS signature.

In Flow, the hash of the threshold signature is the distributed random output. 

## Proof of ownership

A signature scheme can be used to prove the ownership of some secret data _Z_, known only by a limited group of parties. The proof has to be verifiable by any party inside or outside the group. This mechanism is needed in Flow for the [verification stream](/internal/roles/verify).

This can be achieved by using the secret data _Z_ as a source of entropy to generate a key pair following a determinstic public process. The first party to hold _Z_ publishes the generated public key.  
Every party that claims the ownership of _Z_ needs to generate the same key pair and uses the private key to sign a message and publish the signature. The signature is the proof of ownership of _Z_. The signed message needs to be unique to the owner in order to construct a unique proof per owner.  
Any party that challenges the ownership of _Z_ needs to use the unique message and the published public key to verify the signature. 

There is no specific interface for this purpose, the same signature scheme interface described above is enough to achieve this service. 