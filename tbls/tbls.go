package tbls

import (
	"errors"
	"fmt"

	"go.dedis.ch/kyber/v3/pairing"
)

func LagCoef(peer []float64) []float64 {
	lag := []float64{}
	var i int
	n := 3
	for i = 0; i < n; i++ {
		var j int
		var val float64 = 1
		for j = 0; j < n; j++ {
			if i != j {
				p := peer[j] - peer[i]
				p = peer[j] / p
				val = val * p
			}
		}
		lag = append(lag, val)
	}
	return lag
}

// AggregateSignatures combines signatures created using the Sign function
// and Multiplies the shares with the Lagrange Coef.
func AggregateSignatures(suite pairing.Suite, sigs [][]byte, peerList []float64) ([]byte, error) {
	if len(sigs) == 0 || len(peerList) == 0 {
		return nil, errors.New("empty signature list")
	}
	sig := suite.G1().Point()
	l := LagCoef(peerList)
	for i, sigBytes := range sigs {
		sigToAdd := suite.G1().Point()
		if err := sigToAdd.UnmarshalBinary(sigBytes); err != nil {
			fmt.Println("Test")
			return nil, err
		}
		lag := suite.G2().Scalar()
		lag = lag.SetInt64(int64(l[i]))
		sigToAdd.Mul(lag, sigToAdd)
		sig.Add(sig, sigToAdd)
	}
	return sig.MarshalBinary()
}
