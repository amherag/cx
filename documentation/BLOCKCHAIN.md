# Introduction to CX chains

CX chains are CX programs that run on the Skycoin blockchain. A CX chain works by separating a CX program into two parts: its program state and transaction code. The program state is what is stored on the blockchain, and it is updated every time a transaction is run. Consider the code below:

```
package main

func main() {

}
```

One could think that this CX program represents an empty program state, but this is not the case. The program state consists of the following parts: a code segment, a stack segment, a data segment and a heap segment; just like any CX program. In the case of the example above, it would represent a code segment containing a single package with a single function. As the `main` function is empty, nothing will run and the stack and heap segments will be empty. Furthermore, as we do not have any global variables, the data segment will be empty too. Let's have a look at the following CX program:

```
package main

var name str = "Richard"

func main() {

}
```

In this case, we have a code segment similar to the previous example and we have two pieces of data in the data segment: a global variable and the string "Richard" (and the global pointing to that string).

In order to store a CX program on the Skycoin blockchain (which results in a "CX chain") we need to serialize the program. To serialize a program all the packages, structs, functions, globals, etc. of a CX program are converted to byte slices (check `cx/serialize.go` if you are curious), as well as the stack, the data and heap memory segments. As a consequent, if a program with many packages, functions, etc. is used to initiate a CX chain, the program state stored on the blockchain will be big, even if the program is not creating any heap objects or using deep function calls that fill the stack. Just to make it clearer, the following code would also result in a very big program state:

```
package main

var foo [1000][1000]i32

func main() {

}
```

It is important to bear this in mind, as there are hardcoded limits to how big the program state of a CX chain can be, as well as how big a transaction can be.

We now know how and where the program state of a CX chain is going to be stored, but we also need to know how to use it. The program state will never be executed or mutated by itself. In order to modify its state, we need to run a transaction against it. Consider the following pieces of code:

```
package bc

func PlusOne(num i32) (res i32) {
	res = num + 1
}

package main

func main() {

}
```

```
package main
import "bc"

func main() {
	var num i32
	num = bc.PlusOne(5)
	i32.print(num)
}
```

The first code snippet is our "blockchain code" and the second snippet is our "transaction code". As can be seen in the blockchain code, there are two packages: `bc` and `main`. Then, if we pay attention to the transaction code, we can see that its `main` package is importing `bc`. The blockchain code (which will become the CX chain's program state) can be seen as a code library that can store mutable state, and the transaction code can be seen as a program that is importing the blockchain code -- which is exactly what is happening.

"But that's not useful, I can just create another file with that code and import it" you might be thinking. Remember, you can also store state and that state can be mutated:

```
package bc

var numTransactions i32

func LogTransaction() {
	numTransactions++
	i32.print(numTransactions)
}

package main

func main() {

}
```

```
package main
import "bc"

func main() {
	bc.LogTransaction()
}
```

In the example above, we can see that we have a `numTransactions` global variable, which is increased by one if `LogTransaction` is called. This simple program can help us track how many times a transaction has been created and broadcasted.

# TL;DR

Assuming you have a working CX installation.

* Use a Linux system, as it has not been tested on Windows.
* Install Skycoin's cli

```
$ cd $GOPATH/src/github.com/skycoin/skycoin/cmd/cli
$ ./install.sh
```

* Run `skycoin-cli addressGen` and write down the result.
* Replace the fields `blockchain_pubkey_str` and `genesis_address_str` with the values generated by `skycoin-cli addressGen` (`public_key` and `address`, respectively)
* Run `cx --blockchain --heap-initial 100 --stack-size 100 examples/blockchain/counter-bc.cx`.
* Write down the blockchain's genesis signature that is shown at the bottom. CX will update fiber.toml's `genesis_signature_str` field with your new genesis signature.
* Start a publisher node (use the data obtained from running the previous commands; you can create bash variables that hold these values for convenience):

```
cx --publisher --genesis-address $GENESIS_ADDRESS \
   --genesis-signature $GENESIS_SIGNATURE \
   --secret-key $SECRET_KEY \
   --public-key $PUBLIC_KEY
```

* Run a peer node (for now, it MUST use port 6001):

```
cx --peer --genesis-address $GENESIS_ADDRESS \
   --port 6001 \
   --genesis-signature $GENESIS_SIGNATURE \
   --public-key $PUBLIC_KEY
```

* Create a wallet (for now... you need to use this address. This will be changed ASAP):
  * I'm including this code I used for tests. It can also be done with `skycoin-cli`.
  * You MUST export the wallet name and address as $WALLET and $ADDRESS, respectively.

```
    ADDRESS="TkyD4wD64UE6M5BkNQA17zaf7Xcg4AufwX"
    SEED="museum nothing practice weird wheel dignity economy attend mask recipe minor dress"
    LABEL="cxcoin"
    CSRF_TOKEN=$(curl -s http://127.0.0.1:6421/api/v1/csrf | jq -r '.csrf_token')
    
    WALLET=$(curl -s -X POST http://127.0.0.1:6421/api/v1/wallet/create \
         -H "X-CSRF-Token: $CSRF_TOKEN" \
         -H "Content-Type: application/x-www-form-urlencoded" \
         -d "seed=$SEED" \
         -d "coin=$LABEL" \
         -d "label=$LABEL")

    echo $WALLET
    WALLET=$(echo $WALLET | jq -r '.meta.filename')
    
    echo $ADDRESS
    echo $WALLET
    
    export ADDRESS
    export WALLET
```

* You can test a transaction (without broadcasting the new program state) with:

```
cx --transaction --heap-initial 100 --stack-size 100 examples/blockchain/counter-txn.cx
```

* And you can broadcast it with:

```
cx --transaction --heap-initial 100 --stack-size 100 examples/blockchain/counter-txn.cx
```

# Limitations, bugs and non-desirable behaviors

* It has only been tested on Linux (Debian)
* You can't send and receive SKY or any other cryptocurrency on Fiber (which is a good thing, as this CX chains are still in their experimental stage).
* You need to wait a few seconds before creating and broadcasting a new transaction.
* We don't have any security mechanism to prevent someone from calling or accessing certain parts of a CX chain's program state.
* The wallet's address that is sending transactions is hard coded at the moment.
* When broadcasting a transaction, it is also run locally. This is a problem if the transaction takes a considerable amount of time as it would be run locally and then the node would run it again.
* We need a way to set a seed for random numbers for the initial program state. This way we ensure determinism in subsequent transactions. Also, this seed should not be able to be changed by any transaction.

# Creating public, secret key and genesis address

```
skycoin-cli addressGen
```

# Hello, world!

```
cx --heap-initial 100 --stack-size 100 --blockchain examples/blockchain/hello-world-bc.cx
```

```
cx --heap-initial 100 --stack-size 100 --transaction examples/blockchain/hello-world-txn.cx 
```

# Blockchain counter

```
cx --heap-initial 100 --stack-size 100 --blockchain examples/blockchain/counter-bc.cx
```

```
cx --heap-initial 100 --stack-size 100 --transaction examples/blockchain/counter-txn.cx 
```

```
cx --heap-initial 100 --stack-size 100 --broadcast examples/blockchain/counter-txn.cx 
```

# Saving factorial state to run later

# Mario video game, save score

# Smart contract

# Paying salaries