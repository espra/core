// Public Domain (-) 2012-2013 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

//= The hash module defines traits and utility methods for hash functions.

// Hash is the trait implemented by all hash functions.
pub trait Hash {
    fn block_size(self) -> int;
    fn digest(self) -> ~[u8];
    fn digest_size(self) -> int;
    fn reset(&mut self);
    fn update(&mut self, msg: &[u8]);
}

// HashUtil is a utility trait that is auto-implemented for all Hash
// implementations.
pub trait HashUtil {
    fn hexdigest(self) -> ~str;
    fn update(&mut self, msg: &str);
}

impl<A: Hash> HashUtil for A {

    fn hexdigest(self) -> ~str {
        let mut d = ~"";
        for vec::each(self.digest()) |b| { d += fmt!("%02x",*b as uint) }
        return d;
    }

    fn update(&mut self, msg: &str) { self.update(str::to_bytes(msg)); }

}
