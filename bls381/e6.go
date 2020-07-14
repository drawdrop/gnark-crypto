// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gurvy/internal/generators DO NOT EDIT

package bls381

// E6 is a degree-three finite field extension of fp2:
// B0 + B1v + B2v^2 where v^3-1,1 is irrep in fp2
type E6 struct {
	B0, B1, B2 E2
}

// Equal returns true if z equals x, fasle otherwise
// TODO can this be deleted?  Should be able to use == operator instead
func (z *E6) Equal(x *E6) bool {
	return z.B0.Equal(&x.B0) && z.B1.Equal(&x.B1) && z.B2.Equal(&x.B2)
}

// SetString sets a E6 elmt from stringf
func (z *E6) SetString(s1, s2, s3, s4, s5, s6 string) *E6 {
	z.B0.SetString(s1, s2)
	z.B1.SetString(s3, s4)
	z.B2.SetString(s5, s6)
	return z
}

// Set Sets a E6 elmt form another E6 elmt
func (z *E6) Set(x *E6) *E6 {
	z.B0 = x.B0
	z.B1 = x.B1
	z.B2 = x.B2
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E6) SetOne() *E6 {
	z.B0.A0.SetOne()
	z.B0.A1.SetZero()
	z.B1.A0.SetZero()
	z.B1.A1.SetZero()
	z.B2.A0.SetZero()
	z.B2.A1.SetZero()
	return z
}

// SetRandom set z to a random elmt
func (z *E6) SetRandom() *E6 {
	z.B0.SetRandom()
	z.B1.SetRandom()
	z.B2.SetRandom()
	return z
}

// ToMont converts to Mont form
func (z *E6) ToMont() *E6 {
	z.B0.ToMont()
	z.B1.ToMont()
	z.B2.ToMont()
	return z
}

// FromMont converts from Mont form
func (z *E6) FromMont() *E6 {
	z.B0.FromMont()
	z.B1.FromMont()
	z.B2.FromMont()
	return z
}

// Add adds two elements of E6
func (z *E6) Add(x, y *E6) *E6 {
	z.B0.Add(&x.B0, &y.B0)
	z.B1.Add(&x.B1, &y.B1)
	z.B2.Add(&x.B2, &y.B2)
	return z
}

// Neg negates the E6 number
func (z *E6) Neg(x *E6) *E6 {
	z.B0.Neg(&z.B0)
	z.B1.Neg(&z.B1)
	z.B2.Neg(&z.B2)
	return z
}

// Sub two elements of E6
func (z *E6) Sub(x, y *E6) *E6 {
	z.B0.Sub(&x.B0, &y.B0)
	z.B1.Sub(&x.B1, &y.B1)
	z.B2.Sub(&x.B2, &y.B2)
	return z
}

// Double doubles an element in E6
func (z *E6) Double(x *E6) *E6 {
	z.B0.Double(&x.B0)
	z.B1.Double(&x.B1)
	z.B2.Double(&x.B2)
	return z
}

// String puts E6 elmt in string form
func (z *E6) String() string {
	return (z.B0.String() + "+(" + z.B1.String() + ")*v+(" + z.B2.String() + ")*v**2")
}

// Mul sets z to the E6-product of x,y, returns z
func (z *E6) Mul(x, y *E6) *E6 {

	// Algorithm 13 from https://eprint.iacr.org/2010/354.pdf
	var rb0, b0, b1, b2, b3, b4 E2
	b0.Mul(&x.B0, &y.B0) // step 1
	b1.Mul(&x.B1, &y.B1) // step 2
	b2.Mul(&x.B2, &y.B2) // step 3

	// step 4
	b3.Add(&x.B1, &x.B2)
	b4.Add(&y.B1, &y.B2)
	rb0.Mul(&b3, &b4).
		SubAssign(&b1).
		SubAssign(&b2)
	{ // begin inline: set rb0 to (&rb0) * (1,1)
		var buf E2
		buf.Set(&rb0)
		rb0.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(rb0).A0 to (&buf.A1) * (-1)
			(&(rb0).A0).Neg(&buf.A1)
		} // end inline: set &(rb0).A0 to (&buf.A1) * (-1)
		rb0.A0.AddAssign(&buf.A0)
	} // end inline: set rb0 to (&rb0) * (1,1)
	rb0.AddAssign(&b0)

	// step 5
	b3.Add(&x.B0, &x.B1)
	b4.Add(&y.B0, &y.B1)
	z.B1.Mul(&b3, &b4).
		SubAssign(&b0).
		SubAssign(&b1)
	{ // begin inline: set b3 to (&b2) * (1,1)
		var buf E2
		buf.Set(&b2)
		b3.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(b3).A0 to (&buf.A1) * (-1)
			(&(b3).A0).Neg(&buf.A1)
		} // end inline: set &(b3).A0 to (&buf.A1) * (-1)
		b3.A0.AddAssign(&buf.A0)
	} // end inline: set b3 to (&b2) * (1,1)
	z.B1.AddAssign(&b3)

	// step 6
	b3.Add(&x.B0, &x.B2)
	b4.Add(&y.B0, &y.B2)
	z.B2.Mul(&b3, &b4).
		SubAssign(&b0).
		SubAssign(&b2).
		AddAssign(&b1)
	z.B0 = rb0
	return z
}

// MulAssign sets z to the E6-product of z,x returns z
func (z *E6) MulAssign(x *E6) *E6 {

	// Algorithm 13 from https://eprint.iacr.org/2010/354.pdf
	var rb0, b0, b1, b2, b3, b4 E2
	b0.Mul(&z.B0, &x.B0) // step 1
	b1.Mul(&z.B1, &x.B1) // step 2
	b2.Mul(&z.B2, &x.B2) // step 3

	// step 4
	b3.Add(&z.B1, &z.B2)
	b4.Add(&x.B1, &x.B2)
	rb0.Mul(&b3, &b4).
		SubAssign(&b1).
		SubAssign(&b2)
	{ // begin inline: set rb0 to (&rb0) * (1,1)
		var buf E2
		buf.Set(&rb0)
		rb0.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(rb0).A0 to (&buf.A1) * (-1)
			(&(rb0).A0).Neg(&buf.A1)
		} // end inline: set &(rb0).A0 to (&buf.A1) * (-1)
		rb0.A0.AddAssign(&buf.A0)
	} // end inline: set rb0 to (&rb0) * (1,1)
	rb0.AddAssign(&b0)

	// step 5
	b3.Add(&z.B0, &z.B1)
	b4.Add(&x.B0, &x.B1)
	z.B1.Mul(&b3, &b4).
		SubAssign(&b0).
		SubAssign(&b1)
	{ // begin inline: set b3 to (&b2) * (1,1)
		var buf E2
		buf.Set(&b2)
		b3.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(b3).A0 to (&buf.A1) * (-1)
			(&(b3).A0).Neg(&buf.A1)
		} // end inline: set &(b3).A0 to (&buf.A1) * (-1)
		b3.A0.AddAssign(&buf.A0)
	} // end inline: set b3 to (&b2) * (1,1)
	z.B1.AddAssign(&b3)

	// step 6
	b3.Add(&z.B0, &z.B2)
	b4.Add(&x.B0, &x.B2)
	z.B2.Mul(&b3, &b4).
		SubAssign(&b0).
		SubAssign(&b2).
		AddAssign(&b1)
	z.B0 = rb0
	return z
}

// Square sets z to the E6-product of x,x, returns z
func (z *E6) Square(x *E6) *E6 {

	// Algorithm 16 from https://eprint.iacr.org/2010/354.pdf
	var b0, b1, b2, b3, b4 E2
	b3.Mul(&x.B0, &x.B1).Double(&b3) // step 1
	b4.Square(&x.B2)                 // step 2

	// step 3
	{ // begin inline: set b0 to (&b4) * (1,1)
		var buf E2
		buf.Set(&b4)
		b0.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(b0).A0 to (&buf.A1) * (-1)
			(&(b0).A0).Neg(&buf.A1)
		} // end inline: set &(b0).A0 to (&buf.A1) * (-1)
		b0.A0.AddAssign(&buf.A0)
	} // end inline: set b0 to (&b4) * (1,1)
	b0.AddAssign(&b3)
	b1.Sub(&b3, &b4)                                  // step 4
	b2.Square(&x.B0)                                  // step 5
	b3.Sub(&x.B0, &x.B1).AddAssign(&x.B2).Square(&b3) // steps 6 and 8
	b4.Mul(&x.B1, &x.B2).Double(&b4)                  // step 7
	// step 9
	{ // begin inline: set z.B0 to (&b4) * (1,1)
		var buf E2
		buf.Set(&b4)
		z.B0.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(z.B0).A0 to (&buf.A1) * (-1)
			(&(z.B0).A0).Neg(&buf.A1)
		} // end inline: set &(z.B0).A0 to (&buf.A1) * (-1)
		z.B0.A0.AddAssign(&buf.A0)
	} // end inline: set z.B0 to (&b4) * (1,1)
	z.B0.AddAssign(&b2)

	// step 10
	z.B2.Add(&b1, &b3).
		AddAssign(&b4).
		SubAssign(&b2)
	z.B1 = b0
	return z
}

// SquareAssign sets z to the E6-product of z,z returns z
func (z *E6) SquareAssign() *E6 {

	// Algorithm 16 from https://eprint.iacr.org/2010/354.pdf
	var b0, b1, b2, b3, b4 E2
	b3.Mul(&z.B0, &z.B1).Double(&b3) // step 1
	b4.Square(&z.B2)                 // step 2

	// step 3
	{ // begin inline: set b0 to (&b4) * (1,1)
		var buf E2
		buf.Set(&b4)
		b0.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(b0).A0 to (&buf.A1) * (-1)
			(&(b0).A0).Neg(&buf.A1)
		} // end inline: set &(b0).A0 to (&buf.A1) * (-1)
		b0.A0.AddAssign(&buf.A0)
	} // end inline: set b0 to (&b4) * (1,1)
	b0.AddAssign(&b3)
	b1.Sub(&b3, &b4)                                  // step 4
	b2.Square(&z.B0)                                  // step 5
	b3.Sub(&z.B0, &z.B1).AddAssign(&z.B2).Square(&b3) // steps 6 and 8
	b4.Mul(&z.B1, &z.B2).Double(&b4)                  // step 7
	// step 9
	{ // begin inline: set z.B0 to (&b4) * (1,1)
		var buf E2
		buf.Set(&b4)
		z.B0.A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(z.B0).A0 to (&buf.A1) * (-1)
			(&(z.B0).A0).Neg(&buf.A1)
		} // end inline: set &(z.B0).A0 to (&buf.A1) * (-1)
		z.B0.A0.AddAssign(&buf.A0)
	} // end inline: set z.B0 to (&b4) * (1,1)
	z.B0.AddAssign(&b2)

	// step 10
	z.B2.Add(&b1, &b3).
		AddAssign(&b4).
		SubAssign(&b2)
	z.B1 = b0
	return z
}

// Inverse an element in E6
func (z *E6) Inverse(x *E6) *E6 {
	// Algorithm 17 from https://eprint.iacr.org/2010/354.pdf
	// step 9 is wrong in the paper!
	// memalloc
	var t [7]E2
	var c [3]E2
	var buf E2
	t[0].Square(&x.B0)     // step 1
	t[1].Square(&x.B1)     // step 2
	t[2].Square(&x.B2)     // step 3
	t[3].Mul(&x.B0, &x.B1) // step 4
	t[4].Mul(&x.B0, &x.B2) // step 5
	t[5].Mul(&x.B1, &x.B2) // step 6
	// step 7
	{ // begin inline: set c[0] to (&t[5]) * (1,1)
		var buf E2
		buf.Set(&t[5])
		c[0].A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(c[0]).A0 to (&buf.A1) * (-1)
			(&(c[0]).A0).Neg(&buf.A1)
		} // end inline: set &(c[0]).A0 to (&buf.A1) * (-1)
		c[0].A0.AddAssign(&buf.A0)
	} // end inline: set c[0] to (&t[5]) * (1,1)
	c[0].Neg(&c[0]).AddAssign(&t[0])
	// step 8
	{ // begin inline: set c[1] to (&t[2]) * (1,1)
		var buf E2
		buf.Set(&t[2])
		c[1].A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(c[1]).A0 to (&buf.A1) * (-1)
			(&(c[1]).A0).Neg(&buf.A1)
		} // end inline: set &(c[1]).A0 to (&buf.A1) * (-1)
		c[1].A0.AddAssign(&buf.A0)
	} // end inline: set c[1] to (&t[2]) * (1,1)
	c[1].SubAssign(&t[3])
	c[2].Sub(&t[1], &t[4]) // step 9 is wrong in 2010/354!
	// steps 10, 11, 12
	t[6].Mul(&x.B2, &c[1])
	buf.Mul(&x.B1, &c[2])
	t[6].AddAssign(&buf)
	{ // begin inline: set t[6] to (&t[6]) * (1,1)
		var buf E2
		buf.Set(&t[6])
		t[6].A1.Add(&buf.A0, &buf.A1)
		{ // begin inline: set &(t[6]).A0 to (&buf.A1) * (-1)
			(&(t[6]).A0).Neg(&buf.A1)
		} // end inline: set &(t[6]).A0 to (&buf.A1) * (-1)
		t[6].A0.AddAssign(&buf.A0)
	} // end inline: set t[6] to (&t[6]) * (1,1)
	buf.Mul(&x.B0, &c[0])
	t[6].AddAssign(&buf)

	t[6].Inverse(&t[6])    // step 13
	z.B0.Mul(&c[0], &t[6]) // step 14
	z.B1.Mul(&c[1], &t[6]) // step 15
	z.B2.Mul(&c[2], &t[6]) // step 16
	return z
}
