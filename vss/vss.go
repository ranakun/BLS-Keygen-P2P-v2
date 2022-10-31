package vss

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"go.dedis.ch/kyber/v3"
)

var secret big.Int //secret key

// isPrime to check if a number is prime or not
func isPrime(val big.Int) bool {
	if val.ProbablyPrime(0) {
		return true
	} else {
		return false
	}
}

//to generate random prime number of given bit size
func GeneratePrime(size int) *big.Int {
	prime, err := rand.Prime(rand.Reader, size)
	if err != nil {
		fmt.Println(err)
	}
	return prime
}

//choosing a random prime value for q for a given size>=128 bits
func choose_q(size int) *big.Int {
	val := GeneratePrime(size)
	return val
}

//below funcction is to find p which is such that (p-1) is divisble by q or we can say that p-1=r*q or p=r*q+1 where r is an integer
//hence we will start with k=1 and run the algo until we find p such that it is prime and of form mentioned above
func find_p(r int64, q *big.Int) big.Int {
	var p big.Int
	for {
		p.Mul(big.NewInt(r), q)  // p=r*q
		p.Add(&p, big.NewInt(1)) //p=p+1
		if isPrime(p) {
			break
		}
		r++
	}
	return p
}

// Generating polynomial and storing coefficients
//The coefficients of the polynomial should be any random number between 1 and q
func Generate_Polynomial_coefficients(k int64, q *big.Int, poly []*big.Int) {
	file, _ := os.Create("./vss/PolynomialCoefficients.txt") //storing coefficients in a text file
	poly[0] = &secret                                        //constant term of the polynomial will be our secret
	file.WriteString(fmt.Sprintf("%d \n", poly[0]))          //writing a0 to the file
	var i int64
	for i = 1; i < k; i++ {
		coeff, _ := rand.Int(rand.Reader, q) //generating random value for coefficients in range 1 and q
		val := coeff.Sign()
		for val == 0 { //checking if coeffiecient zero then calculating till it's non zero
			coeff, _ = rand.Int(rand.Reader, q)
			val = coeff.Sign()
		}
		poly[i] = coeff
		file.WriteString(fmt.Sprintf("%d \n", coeff)) //writing coeffiecient to the text file
	}
	file.Close()
}

//below function is to generate our Generatoor
func Generate_Generator(p *big.Int) *big.Int {
	g, _ := rand.Int(rand.Reader, p) //choosing g as random number greater than  1 and less than p
	var x big.Int
	x.Sub(g, big.NewInt(1))
	val := x.Sign()
	for val == 0 { //checking if Generator==1
		g, _ := rand.Int(rand.Reader, p)
		x.Sub(g, big.NewInt(1))
		val = x.Sign()
	}
	return g
}

// function to calculate the share of polynomial for ith person

func f_of_i(i int64, k int64, poly []*big.Int) *big.Int {
	// data, _ := ioutil.ReadFile("Polynomial_coefficients.txt")
	// fmt.Println(string(data))
	var j int64
	var val, v1 big.Int
	val = *big.NewInt(0)
	for j = 0; j < k; j++ {
		v1.Exp(big.NewInt(i), big.NewInt(j), nil) // calculating x^j
		//fmt.Println(x)
		v1.Mul(&v1, poly[j]) // multiplying coefficient of j with x^j
		val.Add(&val, &v1)   //adding the values from (j-1)*x^(j-1)
	}
	return &val
}
func generate_share(n int64, k int64, poly []*big.Int, share []*big.Int) {
	var i int64
	file, _ := os.Create("./vss/PolynomialShare.txt") //creating file to store the value ofshare
	for i = 1; i <= n; i++ {
		share[i] = f_of_i(i, k, poly) //to calculate f(i)
		//fmt.Println(share[i])
		file.WriteString(fmt.Sprintf("%d \n", share[i])) //writing share to the file
	}
	file.Close()
}

//below function is to calculate the value of g^a %p
func generate_alpha_i(g *big.Int, a big.Int, p *big.Int) *big.Int {
	var val big.Int
	val.Exp(g, &a, p)
	return &val
}

// to generate the alpha of i for 0<=i<=k-1
func generate_alphas(g *big.Int, k int64, poly []*big.Int, alphas []*big.Int, p *big.Int) {
	//var val big.Int
	var i int64
	for i = 0; i < k; i++ {
		alphas[i] = generate_alpha_i(g, *poly[i], p) //function to calculate g^coeffiecient of i %p
		//fmt.Println(poly[i])
	}
}

// below function is to verfiy the share of each indivisual
func Verify_i(i int64, k int64, g *big.Int, share *big.Int, p *big.Int, alphas []*big.Int) bool {
	var v1, v2 big.Int
	v1.Exp(g, share, p) // v1=g^share %p where share=share of ith person
	//fmt.Println(&v1)
	v2 = *big.NewInt(1) //to store product of alpha[i]^(i^j) for all j from 0 to k-1
	var j int64
	//running the below loop till k
	for j = 0; j < k; j++ {
		var val, val1 big.Int
		val = *big.NewInt(0)
		val1 = *big.NewInt(0)
		val1.Exp(big.NewInt(i), big.NewInt(j), p) //val1=i^j
		val.Exp(alphas[j], &val1, p)              //val=alpha[i]^val i.e val=alpha[i]^(i^j)
		//fmt.Println(alphas[j])
		v2.Mul(&v2, &val)
		v2.Mod(&v2, p)
	}
	// fmt.Println(i)
	// fmt.Println(&v1)
	// fmt.Println(&v2)
	var v3 big.Int
	v3.Sub(&v1, &v2) //subtracting v1 and v2 to check if they are equal if equla then v3=0
	d := v3.Sign()
	//if d==0 means that they are equal and hence verfied
	if d == 0 {
		return true
	} else {
		return false
	}
}

func Setup(s kyber.Scalar, p *big.Int, q *big.Int, g *big.Int) ([]*big.Int, []*big.Int) {
	var k int64
	var i int64
	var n int64

	poly := []*big.Int{}
	share := []*big.Int{}
	alphas := []*big.Int{}

	k = 1
	n = 3

	for i = 0; i <= k; i++ {
		poly = append(poly, big.NewInt(0))
	}
	for i = 0; i <= k; i++ {
		alphas = append(alphas, big.NewInt(0))
	}
	for i = 0; i <= n; i++ {
		share = append(share, big.NewInt(0))
	}

	temp, _ := new(big.Int).SetString(fmt.Sprint(s), 16)
	secret.Set(temp)
	Generate_Polynomial_coefficients(k, q, poly)
	generate_alphas(g, k, poly, alphas, p)
	fmt.Println("\n[+] Generating Shares")
	generate_share(n, k, poly, share)
	for i = 1; i <= n; i++ {
		if !Verify_i(i, k, g, share[i], p, alphas) {
			fmt.Println("error verifying for the shareholder ")
			fmt.Println(i)
			break
		}
	}

	return share, alphas
}

// func Setup(s kyber.Scalar) ([]*big.Int, []*big.Int) {
// 	var r int64 = 1
// 	var k int64
// 	var i int64
// 	var q *big.Int
// 	var p big.Int
// 	var n int64
// 	//giving secret some value let's say q-10
// 	// size := 128 //min size =128
// 	poly := []*big.Int{}
// 	share := []*big.Int{}
// 	alphas := []*big.Int{}
// 	k = 3
// 	n = 5
// 	for i = 0; i <= k; i++ {
// 		poly = append(poly, big.NewInt(0))
// 	}
// 	for i = 0; i <= k; i++ {
// 		alphas = append(alphas, big.NewInt(0))
// 	}
// 	for i = 0; i <= n; i++ {
// 		share = append(share, big.NewInt(0))
// 	}
// 	q = choose_q(2048)
// 	p = find_p(r, q)
// 	// secret.Sub(q, big.NewInt(100))
// 	temp, _ := new(big.Int).SetString(fmt.Sprint(s), 16)
// 	secret.Set(temp)
// 	Generate_Polynomial_coefficients(k, q, poly)
// 	g := Generate_Generator(&p)
// 	//fmt.Println(g)
// 	generate_alphas(g, k, poly, alphas, p)
// 	fmt.Println("[+] Generating Shares")
// 	generate_share(n, k, poly, share)
// 	for i = 1; i <= n; i++ {
// 		// if i == n-1 {
// 		// 	share[i].Add(share[i], big.NewInt(2))
// 		// }
// 		if !verify_i(i, k, g, share[i], p, alphas) {
// 			fmt.Println("error verifying for the shareholder ")
// 			fmt.Println(i)
// 			break
// 		}
// 	}

// 	return share, alphas
// }
