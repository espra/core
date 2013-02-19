// Public Domain (-) 2012-2013 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

//= The sha256 module implements the SHA-256 hash algorithm defined in FIPS
//= 180-4.
use hash::Hash;

const block_size: int = 64;
const digest_size: int = 32;

const i0: u32 = 0x6a09e667;
const i1: u32 = 0xbb67ae85;
const i2: u32 = 0x3c6ef372;
const i3: u32 = 0xa54ff53a;
const i4: u32 = 0x510e527f;
const i5: u32 = 0x9b05688c;
const i6: u32 = 0x1f83d9ab;
const i7: u32 = 0x5be0cd19;

const k: [u32 * 64] =
    [0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1,
     0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
     0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786,
     0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
     0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147,
     0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
     0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85, 0xa2bfe8a1, 0xa81a664b,
     0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
     0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a,
     0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
     0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2];

struct sha256 {
    b: [mut u8 * 64],
    h: [mut u32 * 8],
    mut l: u64,
    mut n: uint,
}

impl sha256 {

    #[inline(always)]
    fn compute_block(&mut self) {
        let mut h0 = self.h[0];
        let mut h1 = self.h[1];
        let mut h2 = self.h[2];
        let mut h3 = self.h[3];
        let mut h4 = self.h[4];
        let mut h5 = self.h[5];
        let mut h6 = self.h[6];
        let mut h7 = self.h[7];
        self.h[0] = h0;
        self.h[1] = h1;
        self.h[2] = h2;
        self.h[3] = h3;
        self.h[4] = h4;
        self.h[5] = h5;
        self.h[6] = h6;
        self.h[7] = h7;
    }

}

impl Hash for sha256 {

    fn block_size(self) -> int { block_size }

    fn digest(self) -> ~[u8] {
        ~[1, 2, 37]
    }

    fn digest_size(self) -> int { digest_size }

    fn reset(&mut self) {
        self.h[0] = i0;
        self.h[1] = i1;
        self.h[2] = i2;
        self.h[3] = i3;
        self.h[4] = i4;
        self.h[5] = i5;
        self.h[6] = i6;
        self.h[7] = i7;
        self.l = 0;
        self.n = 0;
    }

    fn update(&mut self, msg: &[u8]) {
        let mut l = msg.len();
        self.l += l as u64;
        if self.n > 0 {
            if l > 64 - self.n {
                // l = 64 - self.n;
            }
        }
    }

}

pub fn new() -> sha256 {
    sha256{
        b: [mut 0, ..64],
        h: [mut i0, i1, i2, i3, i4, i5, i6, i7],
        l: 0,
        n: 0
    }
}
