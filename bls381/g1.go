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

package bls381

import (
	"runtime"

	"github.com/consensys/gurvy/bls381/fp"

	"math/big"
	"sync"

	"github.com/consensys/gurvy/bls381/fr"
	"github.com/consensys/gurvy/utils/debug"
	"github.com/consensys/gurvy/utils/parallel"
)

// G1Jac is a point with fp.Element coordinates
type G1Jac struct {
	X, Y, Z fp.Element
}

// G1Proj point in projective coordinates
type G1Proj struct {
	X, Y, Z fp.Element
}

// G1Affine point in affine coordinates
type G1Affine struct {
	X, Y fp.Element
}

// g1JacExtended parameterized jacobian coordinates (x=X/ZZ, y=Y/ZZZ, ZZ**3=ZZZ**2)
type g1JacExtended struct {
	X, Y, ZZ, ZZZ fp.Element
}

// SetInfinity sets p to O
func (p *g1JacExtended) SetInfinity() *g1JacExtended {
	p.X.SetOne()
	p.Y.SetOne()
	p.ZZ.SetZero()
	p.ZZZ.SetZero()
	return p
}

// ToAffine sets p in affine coords
func (p *g1JacExtended) ToAffine(Q *G1Affine) *G1Affine {
	Q.X.Inverse(&p.ZZ).MulAssign(&p.X)
	Q.Y.Inverse(&p.ZZZ).MulAssign(&p.Y)
	return Q
}

// ToJac sets p in affine coords
func (p *g1JacExtended) ToJac(Q *G1Jac) *G1Jac {
	Q.X.Mul(&p.ZZ, &p.X).MulAssign(&p.ZZ)
	Q.Y.Mul(&p.ZZZ, &p.Y).MulAssign(&p.ZZZ)
	Q.Z.Set(&p.ZZZ)
	return Q
}

// mAdd
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#addition-madd-2008-s
func (p *g1JacExtended) mAdd(a *G1Affine) *g1JacExtended {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.ZZ.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.ZZ.SetOne()
		p.ZZZ.SetOne()
		return p
	}

	var U2, S2, P, R, PP, PPP, Q, Q2, RR, X3, Y3 fp.Element

	// p2: a, p1: p
	U2.Mul(&a.X, &p.ZZ)
	S2.Mul(&a.Y, &p.ZZZ)
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.double(a)
	}
	P.Sub(&U2, &p.X)
	R.Sub(&S2, &p.Y)
	PP.Square(&P)
	PPP.Mul(&P, &PP)
	Q.Mul(&p.X, &PP)
	RR.Square(&R)
	X3.Sub(&RR, &PPP)
	Q2.AddAssign(&Q).AddAssign(&Q)
	p.X.Sub(&X3, &Q2)
	Y3.Sub(&Q, &p.X).MulAssign(&R)
	R.Mul(&p.Y, &PPP)
	p.Y.Sub(&Y3, &R)
	p.ZZ.MulAssign(&PP)
	p.ZZZ.MulAssign(&PPP)

	return p
}

// double point in ZZ coords
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-xyzz.html#doubling-dbl-2008-s-1
func (p *g1JacExtended) double(q *G1Affine) *g1JacExtended {

	var U, S, M, _M, Y3 fp.Element

	U.Double(&q.Y)
	p.ZZ.Square(&U)
	p.ZZZ.Mul(&U, &p.ZZ)
	S.Mul(&q.X, &p.ZZ)
	_M.Square(&q.X)
	M.Double(&_M).
		AddAssign(&_M) // -> + a, but a=0 here
	p.X.Square(&M).
		SubAssign(&S).
		SubAssign(&S)
	Y3.Sub(&S, &p.X).MulAssign(&M)
	U.Mul(&p.ZZZ, &q.Y)
	p.Y.Sub(&Y3, &U)

	return p
}

// Set set p to the provided point
func (p *G1Jac) Set(a *G1Jac) *G1Jac {
	p.X.Set(&a.X)
	p.Y.Set(&a.Y)
	p.Z.Set(&a.Z)
	return p
}

// Equal tests if two points (in Jacobian coordinates) are equal
func (p *G1Jac) Equal(a *G1Jac) bool {

	if p.Z.IsZero() && a.Z.IsZero() {
		return true
	}
	_p := G1Affine{}
	_p.FromJacobian(p)

	_a := G1Affine{}
	_a.FromJacobian(a)

	return _p.X.Equal(&_a.X) && _p.Y.Equal(&_a.Y)
}

// Equal tests if two points (in Affine coordinates) are equal
func (p *G1Affine) Equal(a *G1Affine) bool {
	return p.X.Equal(&a.X) && p.Y.Equal(&a.Y)
}

// Clone returns a copy of self
func (p *G1Jac) Clone() *G1Jac {
	return &G1Jac{
		p.X, p.Y, p.Z,
	}
}

// Neg computes -G
func (p *G1Jac) Neg(a *G1Jac) *G1Jac {
	p.Set(a)
	p.Y.Neg(&a.Y)
	return p
}

// Neg computes -G
func (p *G1Affine) Neg(a *G1Affine) *G1Affine {
	p.X.Set(&a.X)
	p.Y.Neg(&a.Y)
	return p
}

// SubAssign substracts two points on the curve
func (p *G1Jac) SubAssign(a G1Jac) *G1Jac {
	a.Y.Neg(&a.Y)
	p.AddAssign(&a)
	return p
}

// FromJacobian rescale a point in Jacobian coord in z=1 plane
func (p *G1Affine) FromJacobian(p1 *G1Jac) *G1Affine {

	var a, b fp.Element

	if p1.Z.IsZero() {
		p.X.SetZero()
		p.Y.SetZero()
		return p
	}

	a.Inverse(&p1.Z)
	b.Square(&a)
	p.X.Mul(&p1.X, &b)
	p.Y.Mul(&p1.Y, &b).Mul(&p.Y, &a)

	return p
}

// FromJacobian converts a point from Jacobian to projective coordinates
func (p *G1Proj) FromJacobian(Q *G1Jac) *G1Proj {
	// memalloc
	var buf fp.Element
	buf.Square(&Q.Z)

	p.X.Mul(&Q.X, &Q.Z)
	p.Y.Set(&Q.Y)
	p.Z.Mul(&Q.Z, &buf)

	return p
}

func (p *G1Jac) String() string {
	if p.Z.IsZero() {
		return "O"
	}
	_p := G1Affine{}
	_p.FromJacobian(p)
	return "E([" + _p.X.String() + "," + _p.Y.String() + "]),"
}

// FromAffine sets p = Q, p in Jacboian, Q in affine
func (p *G1Jac) FromAffine(Q *G1Affine) *G1Jac {
	if Q.X.IsZero() && Q.Y.IsZero() {
		p.Z.SetZero()
		p.X.SetOne()
		p.Y.SetOne()
		return p
	}
	p.Z.SetOne()
	p.X.Set(&Q.X)
	p.Y.Set(&Q.Y)
	return p
}

func (p *G1Affine) String() string {
	var x, y fp.Element
	x.Set(&p.X)
	y.Set(&p.Y)
	return "E([" + x.String() + "," + y.String() + "]),"
}

// IsInfinity checks if the point is infinity (in affine, it's encoded as (0,0))
func (p *G1Affine) IsInfinity() bool {
	return p.X.IsZero() && p.Y.IsZero()
}

// AddAssign point addition
// https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#addition-add-2007-bl
func (p *G1Jac) AddAssign(a *G1Jac) *G1Jac {

	// p is infinity, return a
	if p.Z.IsZero() {
		p.Set(a)
		return p
	}

	// a is infinity, return p
	if a.Z.IsZero() {
		return p
	}

	var Z1Z1, Z2Z2, U1, U2, S1, S2, H, I, J, r, V fp.Element
	Z1Z1.Square(&a.Z)
	Z2Z2.Square(&p.Z)
	U1.Mul(&a.X, &Z2Z2)
	U2.Mul(&p.X, &Z1Z1)
	S1.Mul(&a.Y, &p.Z).
		MulAssign(&Z2Z2)
	S2.Mul(&p.Y, &a.Z).
		MulAssign(&Z1Z1)

	// if p == a, we double instead
	if U1.Equal(&U2) && S1.Equal(&S2) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &U1)
	I.Double(&H).
		Square(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &S1).Double(&r)
	V.Mul(&U1, &I)
	p.X.Square(&r).
		SubAssign(&J).
		SubAssign(&V).
		SubAssign(&V)
	p.Y.Sub(&V, &p.X).
		MulAssign(&r)
	S1.MulAssign(&J).Double(&S1)
	p.Y.SubAssign(&S1)
	p.Z.AddAssign(&a.Z)
	p.Z.Square(&p.Z).
		SubAssign(&Z1Z1).
		SubAssign(&Z2Z2).
		MulAssign(&H)

	return p
}

// AddMixed point addition
// http://www.hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-0.html#addition-madd-2007-bl
func (p *G1Jac) AddMixed(a *G1Affine) *G1Jac {

	//if a is infinity return p
	if a.X.IsZero() && a.Y.IsZero() {
		return p
	}
	// p is infinity, return a
	if p.Z.IsZero() {
		p.X = a.X
		p.Y = a.Y
		p.Z.SetOne()
		return p
	}

	var Z1Z1, U2, S2, H, HH, I, J, r, V fp.Element
	Z1Z1.Square(&p.Z)
	U2.Mul(&a.X, &Z1Z1)
	S2.Mul(&a.Y, &p.Z).
		MulAssign(&Z1Z1)

	// if p == a, we double instead
	if U2.Equal(&p.X) && S2.Equal(&p.Y) {
		return p.DoubleAssign()
	}

	H.Sub(&U2, &p.X)
	HH.Square(&H)
	I.Double(&HH).Double(&I)
	J.Mul(&H, &I)
	r.Sub(&S2, &p.Y).Double(&r)
	V.Mul(&p.X, &I)
	p.X.Square(&r).
		SubAssign(&J).
		SubAssign(&V).
		SubAssign(&V)
	J.MulAssign(&p.Y).Double(&J)
	p.Y.Sub(&V, &p.X).
		MulAssign(&r)
	p.Y.SubAssign(&J)
	p.Z.AddAssign(&H)
	p.Z.Square(&p.Z).
		SubAssign(&Z1Z1).
		SubAssign(&HH)

	return p
}

// DoubleAssign doubles a point in Jacobian coordinates
// https://hyperelliptic.org/EFD/g1p/auto-shortw-jacobian-3.html#doubling-dbl-2007-bl
func (p *G1Jac) DoubleAssign() *G1Jac {

	if p.Z.IsZero() {
		return p
	}

	var XX, YY, YYYY, ZZ, S, M, T fp.Element

	XX.Square(&p.X)
	YY.Square(&p.Y)
	YYYY.Square(&YY)
	ZZ.Square(&p.Z)
	S.Add(&p.X, &YY)
	S.Square(&S).
		SubAssign(&XX).
		SubAssign(&YYYY).
		Double(&S)
	M.Double(&XX).AddAssign(&XX)
	p.Z.AddAssign(&p.Y).
		Square(&p.Z).
		SubAssign(&YY).
		SubAssign(&ZZ)
	T.Square(&M)
	p.X = T
	T.Double(&S)
	p.X.SubAssign(&T)
	p.Y.Sub(&S, &p.X).
		MulAssign(&M)
	YYYY.Double(&YYYY).Double(&YYYY).Double(&YYYY)
	p.Y.SubAssign(&YYYY)

	return p
}

// doubleandadd algo for exponentiation
func (p *G1Jac) _doubleandadd(a *G1Affine, s big.Int) *G1Jac {

	var res G1Jac
	res.Set(&g1Infinity)
	b := s.Bytes()
	for i := range b {
		w := b[i]
		mask := byte(0x80)
		for j := 0; j < 8; j++ {
			res.DoubleAssign()
			if (w&mask)>>(7-j) != 0 {
				res.AddMixed(a)
			}
			mask = mask >> 1
		}
	}
	p.Set(&res)

	return p
}

// ScalarMul multiplies a by scalar
// algorithm: a special case of Pippenger described by Bootle:
// https://jbootle.github.io/Misc/pippenger.pdf
func (p *G1Jac) ScalarMul(a *G1Jac, scalar fr.Element) *G1Jac {
	// see MultiExp and pippenger documentation for more details about these constants / variables
	const s = 4
	const b = s
	const TSize = (1 << b) - 1
	var T [TSize]G1Jac
	computeT := func(T []G1Jac, t0 *G1Jac) {
		T[0].Set(t0)
		for j := 1; j < (1<<b)-1; j = j + 2 {
			T[j].Set(&T[j/2]).DoubleAssign()
			T[j+1].Set(&T[(j+1)/2]).AddAssign(&T[j/2])
		}
	}
	return p.pippenger([]G1Jac{*a}, []fr.Element{scalar}, s, b, T[:], computeT)
}

// ScalarMulByGen multiplies curve.g1Gen by scalar
// algorithm: a special case of Pippenger described by Bootle:
// https://jbootle.github.io/Misc/pippenger.pdf
func (p *G1Jac) ScalarMulByGen(scalar fr.Element) *G1Jac {
	computeT := func(T []G1Jac, t0 *G1Jac) {}
	return p.pippenger([]G1Jac{g1Gen}, []fr.Element{scalar}, sGen, bGen, tGenG1[:], computeT)
}

// WindowedMultiExp set p = scalars[0]*points[0] + ... + scalars[n]*points[n]
// assume: scalars in non-Montgomery form!
// assume: len(points)==len(scalars)>0, len(scalars[i]) equal for all i
// algorithm: a special case of Pippenger described by Bootle:
// https://jbootle.github.io/Misc/pippenger.pdf
// uses all availables runtime.NumCPU()
func (p *G1Jac) WindowedMultiExp(points []G1Jac, scalars []fr.Element) *G1Jac {
	var lock sync.Mutex
	parallel.Execute(0, len(points), func(start, end int) {
		var t G1Jac
		t.multiExp(points[start:end], scalars[start:end])
		lock.Lock()
		p.AddAssign(&t)
		lock.Unlock()
	}, false)
	return p
}

// multiExp set p = scalars[0]*points[0] + ... + scalars[n]*points[n]
// assume: scalars in non-Montgomery form!
// assume: len(points)==len(scalars)>0, len(scalars[i]) equal for all i
// algorithm: a special case of Pippenger described by Bootle:
// https://jbootle.github.io/Misc/pippenger.pdf
func (p *G1Jac) multiExp(points []G1Jac, scalars []fr.Element) *G1Jac {
	const s = 4 // s from Bootle, we choose s divisible by scalar bit length
	const b = s // b from Bootle, we choose b equal to s
	// WARNING! This code breaks if you switch to b!=s
	// Because we chose b=s, each set S_i from Bootle is simply the set of points[i]^{2^j} for each j in [0:s]
	// This choice allows for simpler code
	// If you want to use b!=s then the S_i from Bootle are different
	const TSize = (1 << b) - 1 // TSize is size of T_i sets from Bootle, equal to 2^b - 1
	// Store only one set T_i at a time---don't store them all!
	var T [TSize]G1Jac // a set T_i from Bootle, the set of g^j for j in [1:2^b] for some choice of g
	computeT := func(T []G1Jac, t0 *G1Jac) {
		T[0].Set(t0)
		for j := 1; j < (1<<b)-1; j = j + 2 {
			T[j].Set(&T[j/2]).DoubleAssign()
			T[j+1].Set(&T[(j+1)/2]).AddAssign(&T[j/2])
		}
	}
	return p.pippenger(points, scalars, s, b, T[:], computeT)
}

// algorithm: a special case of Pippenger described by Bootle:
// https://jbootle.github.io/Misc/pippenger.pdf
func (p *G1Jac) pippenger(points []G1Jac, scalars []fr.Element, s, b uint64, T []G1Jac, computeT func(T []G1Jac, t0 *G1Jac)) *G1Jac {
	var t, selectorIndex, ks int
	var selectorMask, selectorShift, selector uint64

	t = fr.ElementLimbs * 64 / int(s) // t from Bootle, equal to (scalar bit length) / s
	selectorMask = (1 << b) - 1       // low b bits are 1
	morePoints := make([]G1Jac, t)    // morePoints is the set of G'_k points from Bootle
	for k := 0; k < t; k++ {
		morePoints[k].Set(&g1Infinity)
	}
	for i := 0; i < len(points); i++ {
		// compute the set T_i from Bootle: all possible combinations of elements from S_i from Bootle
		computeT(T, &points[i])
		// for each morePoints: find the right T element and add it
		for k := 0; k < t; k++ {
			ks = k * int(s)
			selectorIndex = ks / 64
			selectorShift = uint64(ks - (selectorIndex * 64))
			selector = (scalars[i][selectorIndex] & (selectorMask << selectorShift)) >> selectorShift
			if selector != 0 {
				morePoints[k].AddAssign(&T[selector-1])
			}
		}
	}
	// combine morePoints to get the final result
	p.Set(&morePoints[t-1])
	for k := t - 2; k >= 0; k-- {
		for j := uint64(0); j < s; j++ {
			p.DoubleAssign()
		}
		p.AddAssign(&morePoints[k])
	}
	return p
}

// MultiExp complexity O(n)
func (p *G1Jac) MultiExp(points []G1Affine, scalars []fr.Element) chan G1Jac {
	nbPoints := len(points)
	debug.Assert(nbPoints == len(scalars))

	chRes := make(chan G1Jac, 1)

	// under 50 points, the windowed multi exp performs better
	const minPoints = 50
	if nbPoints <= minPoints {
		_points := make([]G1Jac, len(points))
		for i := 0; i < len(points); i++ {
			_points[i].FromAffine(&points[i])
		}
		go func() {
			p.WindowedMultiExp(_points, scalars)
			chRes <- *p
		}()
		return chRes
	}

	// empirical values
	var nbChunks, chunkSize int
	var mask uint64
	if nbPoints <= 10000 {
		chunkSize = 8
	} else if nbPoints <= 80000 {
		chunkSize = 11
	} else if nbPoints <= 400000 {
		chunkSize = 13
	} else if nbPoints <= 800000 {
		chunkSize = 14
	} else {
		chunkSize = 16
	}

	const sizeScalar = fr.ElementLimbs * 64

	var bitsForTask [][]int
	if sizeScalar%chunkSize == 0 {
		counter := sizeScalar - 1
		nbChunks = sizeScalar / chunkSize
		bitsForTask = make([][]int, nbChunks)
		for i := 0; i < nbChunks; i++ {
			bitsForTask[i] = make([]int, chunkSize)
			for j := 0; j < chunkSize; j++ {
				bitsForTask[i][j] = counter
				counter--
			}
		}
	} else {
		counter := sizeScalar - 1
		nbChunks = sizeScalar/chunkSize + 1
		bitsForTask = make([][]int, nbChunks)
		for i := 0; i < nbChunks; i++ {
			if i < nbChunks-1 {
				bitsForTask[i] = make([]int, chunkSize)
			} else {
				bitsForTask[i] = make([]int, sizeScalar%chunkSize)
			}
			for j := 0; j < chunkSize && counter >= 0; j++ {
				bitsForTask[i][j] = counter
				counter--
			}
		}
	}

	accumulators := make([]G1Jac, nbChunks)
	chIndices := make([]chan struct{}, nbChunks)
	chPoints := make([]chan struct{}, nbChunks)
	for i := 0; i < nbChunks; i++ {
		chIndices[i] = make(chan struct{}, 1)
		chPoints[i] = make(chan struct{}, 1)
	}

	mask = (1 << chunkSize) - 1
	nbPointsPerSlots := nbPoints / int(mask)
	// [][] is more efficient than [][][] for storage, elements are accessed via i*nbChunks+k
	indices := make([][]int, int(mask)*nbChunks)
	for i := 0; i < int(mask)*nbChunks; i++ {
		indices[i] = make([]int, 0, nbPointsPerSlots)
	}

	// if chunkSize=8, nbChunks=32 (the scalars are chunkSize*nbChunks bits long)
	// for each 32 chunk, there is a list of 2**8=256 list of indices
	// for the i-th chunk, accumulateIndices stores in the k-th list all the indices of points
	// for which the i-th chunk of 8 bits is equal to k
	accumulateIndices := func(cpuID, nbTasks, n int) {
		for i := 0; i < nbTasks; i++ {
			task := cpuID + i*n
			idx := task*int(mask) - 1
			for j := 0; j < nbPoints; j++ {
				val := 0
				for k := 0; k < len(bitsForTask[task]); k++ {
					val = val << 1
					c := bitsForTask[task][k] / int(64)
					o := bitsForTask[task][k] % int(64)
					b := (scalars[j][c] >> o) & 1
					val += int(b)
				}
				if val != 0 {
					indices[idx+int(val)] = append(indices[idx+int(val)], j)
				}
			}
			chIndices[task] <- struct{}{}
			close(chIndices[task])
		}
	}

	// if chunkSize=8, nbChunks=32 (the scalars are chunkSize*nbChunks bits long)
	// for each chunk, sum up elements in index 0, add to current result, sum up elements
	// in index 1, add to current result, etc, up to 255=2**8-1
	accumulatePoints := func(cpuID, nbTasks, n int) {
		for i := 0; i < nbTasks; i++ {
			var tmp g1JacExtended
			var _tmp G1Jac
			task := cpuID + i*n

			// init points
			tmp.SetInfinity()
			accumulators[task].Set(&g1Infinity)

			// wait for indices to be ready
			<-chIndices[task]

			for j := int(mask - 1); j >= 0; j-- {
				for _, k := range indices[task*int(mask)+j] {
					tmp.mAdd(&points[k])
				}
				tmp.ToJac(&_tmp)
				accumulators[task].AddAssign(&_tmp)
			}
			chPoints[task] <- struct{}{}
			close(chPoints[task])
		}
	}

	// double and add algo to collect all small reductions
	reduce := func() {
		var res G1Jac
		res.Set(&g1Infinity)
		for i := 0; i < nbChunks; i++ {
			for j := 0; j < len(bitsForTask[i]); j++ {
				res.DoubleAssign()
			}
			<-chPoints[i]
			res.AddAssign(&accumulators[i])
		}
		p.Set(&res)
		chRes <- *p
	}

	nbCpus := runtime.NumCPU()
	nbTasksPerCpus := nbChunks / nbCpus
	remainingTasks := nbChunks % nbCpus
	for i := 0; i < nbCpus; i++ {
		if remainingTasks > 0 {
			go accumulateIndices(i, nbTasksPerCpus+1, nbCpus)
			go accumulatePoints(i, nbTasksPerCpus+1, nbCpus)
			remainingTasks--
		} else {
			go accumulateIndices(i, nbTasksPerCpus, nbCpus)
			go accumulatePoints(i, nbTasksPerCpus, nbCpus)
		}
	}

	go reduce()

	return chRes
}
