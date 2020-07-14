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

import (
	"github.com/consensys/gurvy/bls381/fp"
	"math/bits"
)

// FinalExponentiation computes the final expo x**(p**6-1)(p**2+1)(p**4 - p**2 +1)/r
func (curve *Curve) FinalExponentiation(z *PairingResult, _z ...*PairingResult) PairingResult {
	var result PairingResult
	result.Set(z)

	// if additional parameters are provided, multiply them into z
	for _, e := range _z {
		result.Mul(&result, e)
	}

	result.FinalExponentiation(&result)

	return result
}

// FinalExponentiation sets z to the final expo x**((p**12 - 1)/r), returns z
func (z *PairingResult) FinalExponentiation(x *PairingResult) *PairingResult {
	// For BLS curves use Section 3 of https://eprint.iacr.org/2016/130.pdf; "hard part" is Algorithm 1 of https://eprint.iacr.org/2016/130.pdf
	var result PairingResult
	result.Set(x)

	// memalloc
	var t [6]PairingResult

	// buf = x**(p^6-1)
	t[0].FrobeniusCube(&result).
		FrobeniusCube(&t[0])

	result.Inverse(&result)
	t[0].Mul(&t[0], &result)

	// x = (x**(p^6-1)) ^(p^2+1)
	result.FrobeniusSquare(&t[0]).
		Mul(&result, &t[0])

	// hard part (up to permutation)
	// performs the hard part of the final expo
	// Algorithm 1 of https://eprint.iacr.org/2016/130.pdf
	// The result is the same as p**4-p**2+1/r, but up to permutation (it's 3* (p**4 -p**2 +1 /r)), ok since r=1 mod 3)

	t[0].InverseUnitary(&result).Square(&t[0])
	t[5].Expt(&result)
	t[1].CyclotomicSquare(&t[5])
	t[3].Mul(&t[0], &t[5])

	t[0].Expt(&t[3])
	t[2].Expt(&t[0])
	t[4].Expt(&t[2])

	t[4].Mul(&t[1], &t[4])
	t[1].Expt(&t[4])
	t[3].InverseUnitary(&t[3])
	t[1].Mul(&t[3], &t[1])
	t[1].Mul(&t[1], &result)

	t[0].Mul(&t[0], &result)
	t[0].FrobeniusCube(&t[0])

	t[3].InverseUnitary(&result)
	t[4].Mul(&t[3], &t[4])
	t[4].Frobenius(&t[4])

	t[5].Mul(&t[2], &t[5])
	t[5].FrobeniusSquare(&t[5])

	t[5].Mul(&t[5], &t[0])
	t[5].Mul(&t[5], &t[4])
	t[5].Mul(&t[5], &t[1])

	result.Set(&t[5])

	z.Set(&result)
	return z
}

// MillerLoop Miller loop
func (curve *Curve) MillerLoop(P G1Affine, Q G2Affine, result *PairingResult) *PairingResult {

	// init result
	result.SetOne()

	if P.IsInfinity() || Q.IsInfinity() {
		return result
	}

	// the line goes through QCur and QNext
	var QCur, QNext, QNextNeg G2Jac
	var QNeg G2Affine

	// Stores -Q
	QNeg.Neg(&Q)

	// init QCur with Q
	QCur.FromAffine(&Q)

	var lEval lineEvalRes

	// Miller loop
	for i := len(curve.loopCounter) - 2; i >= 0; i-- {
		QNext.Set(&QCur)
		QNext.DoubleAssign()
		QNextNeg.Neg(&QNext)

		result.Square(result)

		// evaluates line though Qcur,2Qcur at P
		lineEvalJac(QCur, QNextNeg, &P, &lEval)
		lEval.mulAssign(result)

		if curve.loopCounter[i] == 1 {
			// evaluates line through 2Qcur, Q at P
			lineEvalAffine(QNext, Q, &P, &lEval)
			lEval.mulAssign(result)

			QNext.AddMixed(&Q)

		} else if curve.loopCounter[i] == -1 {
			// evaluates line through 2Qcur, -Q at P
			lineEvalAffine(QNext, QNeg, &P, &lEval)
			lEval.mulAssign(result)

			QNext.AddMixed(&QNeg)
		}
		QCur.Set(&QNext)
	}

	return result
}

// lineEval computes the evaluation of the line through Q, R (on the twist) at P
// Q, R are in jacobian coordinates
// The case in which Q=R=Infinity is not handled as this doesn't happen in the SNARK pairing
func lineEvalJac(Q, R G2Jac, P *G1Affine, result *lineEvalRes) {
	// converts _Q and _R to projective coords
	var _Q, _R G2Proj
	_Q.FromJacobian(&Q)
	_R.FromJacobian(&R)

	// line eq: w^3*(_Qy_Rz-_Qz_Ry)x +  w^2*(_Qz_Rx - _Qx_Rz)y + w^5*(_Qx_Ry-_Qy_Rxz)
	// result.r1 = _Qy_Rz-_Qz_Ry
	// result.r0 = _Qz_Rx - _Qx_Rz
	// result.r2 = _Qx_Ry-_Qy_Rxz

	result.r1.Mul(&_Q.Y, &_R.Z)
	result.r0.Mul(&_Q.Z, &_R.X)
	result.r2.Mul(&_Q.X, &_R.Y)

	_Q.Z.Mul(&_Q.Z, &_R.Y)
	_Q.X.Mul(&_Q.X, &_R.Z)
	_Q.Y.Mul(&_Q.Y, &_R.X)

	result.r1.Sub(&result.r1, &_Q.Z)
	result.r0.Sub(&result.r0, &_Q.X)
	result.r2.Sub(&result.r2, &_Q.Y)

	// multiply P.Z by coeffs[2] in case P is infinity
	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
	//result.r2.MulByElement(&result.r2, &P.Z)
}

// Same as above but R is in affine coords
func lineEvalAffine(Q G2Jac, R G2Affine, P *G1Affine, result *lineEvalRes) {

	// converts Q and R to projective coords
	var _Q G2Proj
	_Q.FromJacobian(&Q)

	// line eq: w^3*(QyRz-QzRy)x +  w^2*(QzRx - QxRz)y + w^5*(QxRy-QyRxz)
	// result.r1 = QyRz-QzRy
	// result.r0 = QzRx - QxRz
	// result.r2 = QxRy-QyRxz

	result.r1.Set(&_Q.Y)
	result.r0.Mul(&_Q.Z, &R.X)
	result.r2.Mul(&_Q.X, &R.Y)

	_Q.Z.Mul(&_Q.Z, &R.Y)
	_Q.Y.Mul(&_Q.Y, &R.X)

	result.r1.Sub(&result.r1, &_Q.Z)
	result.r0.Sub(&result.r0, &_Q.X)
	result.r2.Sub(&result.r2, &_Q.Y)

	// multiply P.Z by coeffs[2] in case P is infinity
	result.r1.MulByElement(&result.r1, &P.X)
	result.r0.MulByElement(&result.r0, &P.Y)
	// result.r2.MulByElement(&result.r2, &P.Z)
}

type lineEvalRes struct {
	r0 G2CoordType // c0.b1
	r1 G2CoordType // c1.b1
	r2 G2CoordType // c1.b2
}

func (l *lineEvalRes) mulAssign(z *PairingResult) *PairingResult {

	var a, b, c PairingResult
	a.MulByVWNRInv(z, &l.r1)
	b.MulByV2NRInv(z, &l.r0)
	c.MulByWNRInv(z, &l.r2)
	z.Add(&a, &b).Add(z, &c)

	return z
}

// MulByV2NRInv set z to x*(y*v^2*(1,1)^{-1}) and return z
// here y*v^2 means the PairingResult element with C0.B2=y and all other components 0
func (z *PairingResult) MulByV2NRInv(x *PairingResult, y *G2CoordType) *PairingResult {
	var result PairingResult
	var yNRInv G2CoordType
	yNRInv.mulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C0.B1, y)
	result.C0.B1.Mul(&x.C0.B2, y)
	result.C0.B2.Mul(&x.C0.B0, &yNRInv)

	result.C1.B0.Mul(&x.C1.B1, y)
	result.C1.B1.Mul(&x.C1.B2, y)
	result.C1.B2.Mul(&x.C1.B0, &yNRInv)

	z.Set(&result)
	return z
}

// MulByVWNRInv set z to x*(y*v*w*(1,1)^{-1}) and return z
// here y*v*w means the PairingResult element with C1.B1=y and all other components 0
func (z *PairingResult) MulByVWNRInv(x *PairingResult, y *G2CoordType) *PairingResult {
	var result PairingResult
	var yNRInv G2CoordType
	yNRInv.mulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C1.B1, y)
	result.C0.B1.Mul(&x.C1.B2, y)
	result.C0.B2.Mul(&x.C1.B0, &yNRInv)

	result.C1.B0.Mul(&x.C0.B2, y)
	result.C1.B1.Mul(&x.C0.B0, &yNRInv)
	result.C1.B2.Mul(&x.C0.B1, &yNRInv)

	z.Set(&result)
	return z
}

// MulByWNRInv set z to x*(y*w*(1,1)^{-1}) and return z
// here y*w means the PairingResult element with C1.B0=y and all other components 0
func (z *PairingResult) MulByWNRInv(x *PairingResult, y *G2CoordType) *PairingResult {
	var result PairingResult
	var yNRInv G2CoordType
	yNRInv.mulByNonResidueInv(y)

	result.C0.B0.Mul(&x.C1.B2, y)
	result.C0.B1.Mul(&x.C1.B0, &yNRInv)
	result.C0.B2.Mul(&x.C1.B1, &yNRInv)

	result.C1.B0.Mul(&x.C0.B0, &yNRInv)
	result.C1.B1.Mul(&x.C0.B1, &yNRInv)
	result.C1.B2.Mul(&x.C0.B2, &yNRInv)

	z.Set(&result)
	return z
}

// mulByNonResidueInv set z to x * (1,1)^{-1} and return z
func (z *G2CoordType) mulByNonResidueInv(x *G2CoordType) *G2CoordType {
	{ // begin inline: set z to x * (1,1)^{-1}
		// z.A0 = (x.A0 + x.A1)/2
		// z.A1 = (x.A1 - x.A0)/2
		buf := *x
		z.A0.Add(&buf.A0, &buf.A1)
		z.A1.Sub(&buf.A1, &buf.A0)
		twoInv := fp.Element{
			1730508156817200468,
			9606178027640717313,
			7150789853162776431,
			7936136305760253186,
			15245073033536294050,
			1728177566264616342,
		}
		z.A0.MulAssign(&twoInv)
		z.A1.MulAssign(&twoInv)
	} // end inline: set z to x * (1,1)^{-1}
	return z
}

const tAbsVal uint64 = 15132376222941642752 // negative

// Expt set z to x^t in PairingResult and return z
func (z *PairingResult) Expt(x *PairingResult) *PairingResult {

	var result PairingResult
	result.Set(x)

	l := bits.Len64(tAbsVal) - 2
	for i := l; i >= 0; i-- {
		result.CyclotomicSquare(&result)
		if tAbsVal&(1<<uint(i)) != 0 {
			result.Mul(&result, x)
		}
	}
	result.Conjugate(&result) // because tAbsVal is negative

	z.Set(&result)
	return z
}
