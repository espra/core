// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

//= Package hash defines traits and utility methods for hash functions.

// Hash is the trait implemented by all hash functions.
pub trait Hash {
    fn block_size() -> int;
    fn digest() -> ~[byte];
    fn digest_size() -> int;
    fn reset();
    fn update(msg: & [byte]);
}

// HashUtil is a utility trait that is auto-implemented for all Hash
// implementations.
pub trait HashUtil {
    fn hexdigest() -> ~str;
    fn update(msg: & str);
}

impl <A: Hash> A: HashUtil {

    fn hexdigest() -> ~str {
        let mut d = ~"";
        for vec::each(self.digest()) |b| { d += fmt!("%02x",*b as uint) }
        return d;
    }

    fn update(msg: & str) { self.update(str::to_bytes(msg)); }

}
