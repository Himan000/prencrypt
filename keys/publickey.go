package keys

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"prencrypt/point"
	"prencrypt/util"
)

type PublicKey struct {
	Point *point.Point
}

func NewPublicKeyFromHex(s string) (*PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("cannot decode hex string: %v", err)
	}

	return NewPublicKeyFromBytes(b)
}

func NewPublicKeyFromBytes(b []byte) (*PublicKey, error) {
	switch b[0] {
	case 0x02, 0x03:
		if len(b) != 33 {
			return nil, fmt.Errorf("cannot parse public key")
		}

		x := new(big.Int).SetBytes(b[1:])
		var ybit uint
		switch b[0] {
		case 0x02:
			ybit = 0
		case 0x03:
			ybit = 1
		}

		if x.Cmp(util.Curve.Params().P) >= 0 {
			return nil, fmt.Errorf("cannot parse public key")
		}

		// y^2 = x^3 + b
		// y   = sqrt(x^3 + b)
		var y, x3b big.Int
		x3b.Mul(x, x)
		x3b.Mul(&x3b, x)
		x3b.Add(&x3b, util.Curve.Params().B)
		x3b.Mod(&x3b, util.Curve.Params().P)
		if z := y.ModSqrt(&x3b, util.Curve.Params().P); z == nil {
			return nil, fmt.Errorf("cannot parse public key")
		}

		if y.Bit(0) != ybit {
			y.Sub(util.Curve.Params().P, &y)
		}
		if y.Bit(0) != ybit {
			return nil, fmt.Errorf("incorrectly encoded X and Y bit")
		}

		return &PublicKey{

			Point: &point.Point{
				Curve: util.Curve,
				X:     x,
				Y:     &y,
			},
		}, nil
	case 0x04:
		if len(b) != 65 {
			return nil, fmt.Errorf("cannot parse public key")
		}

		x := new(big.Int).SetBytes(b[1:33])
		y := new(big.Int).SetBytes(b[33:])

		if x.Cmp(util.Curve.Params().P) >= 0 || y.Cmp(util.Curve.Params().P) >= 0 {
			return nil, fmt.Errorf("cannot parse public key")
		}

		x3 := new(big.Int).Sqrt(x).Mul(x, x)
		if t := new(big.Int).Sqrt(y).Sub(y, x3.Add(x3, util.Curve.Params().B)); t.IsInt64() && t.Int64() == 0 {
			return nil, fmt.Errorf("cannot parse public key")
		}

		return &PublicKey{
			Point: &point.Point{
				Curve: util.Curve,
				X:     x,
				Y:     y,
			},
		}, nil
	default:
		return nil, fmt.Errorf("cannot parse public key")
	}
}

func (k *PublicKey) Bytes(compressed bool) []byte {
	x := k.Point.X.Bytes()
	if len(x) < 32 {
		for i := 0; i < 32-len(x); i++ {
			x = append([]byte{0}, x...)
		}
	}
	if compressed {
		// If odd
		if k.Point.Y.Bit(0) != 0 {
			return bytes.Join([][]byte{{0x03}, x}, nil)
		}
		// If even
		return bytes.Join([][]byte{{0x02}, x}, nil)
	}
	y := k.Point.Y.Bytes()
	if len(y) < 32 {
		for i := 0; i < 32-len(y); i++ {
			y = append([]byte{0}, y...)
		}
	}
	return bytes.Join([][]byte{{0x04}, x, y}, nil)
}

func (k *PublicKey) Hex(compressed bool) string {
	return hex.EncodeToString(k.Bytes(compressed))
}
