package cxcore

// This package should mimic Golang's net/http package: https://golang.org/pkg/net/http/

import (
	"net/http"
)

// op_http_ListenAndServe this is just a mock example that explains the general process
// for creating a "wrapper" library for CX. I call them wrappers because these functions
// usually just parse a CX expression to extract its inputs, then sends these inputs to
// Golang functions, then parse the results of these Golang functions back to CX outputs
// and writes the outputs back to CX memory.
func op_http_ListenAndServe(prgrm *CXProgram) {
	// We get the expression that was called in the CX source code.
	// In this case, it could be `http.ListenAndServe(":8080")`
	// Golang's version of `ListenAndServe` requires two inputs: a port and a handler.
	// But we can manage it with a `nil` handler, so we can ignore that input.
	// If we decided not to ignore it, we'd need to make a conversion from Golang handler
	// structure -> CX handler structure, which is doable, but we lack an easy way to do it.
	expr := prgrm.GetExpr()

	// We get the frame pointer. This will just tell us where in the stack segment we can start writing.
	fp := prgrm.GetFramePointer()

	// We extract the first input, which is ":8080"
	inp1 := expr.Inputs[0]

	// We extract the string ":8080" from CX memory using `ReadStr`.
	// `ReadStr` first reads from the stack, using `fp` + the relative offset of `inp1`
	// and then uses the address located in there to arrive to the final string, which
	// could be either located in the data segment or in the heap segment. You shouldn't
	// need to worry about this process, but it may be good to know.
	port := ReadStr(fp, inp1)

	// We do the actual Golang `ListenAndServe` call, using the port extracted from CX memory.
	err := http.ListenAndServe(port, nil)

	// In this case I decided to just panic if there's an `err`. The best way would be to
	// return the error to the CX program.
	if err != nil {
		panic(err)
	}

	// For example, if we decided to return the error as a CX `str`, we could have done it this way:
	// writeString(fp, err.String(), out1)
	// In this case, `writeString` handles everything for us. `writeString` is pretty convenient, because it
	// does all of this:
	// it creates a heap object with the actual error string in it, gets the address of that heap object
	// and writes the address to the stack, where the output of `ListenAndServe` should go.

	// If you need to write non-heap objects to the stack (i.e., i32, i64, bool, f32, f64, and **not** slices,
	// pointers or strings), you'd need to do something like this:
	// We serialize whatever you want to be serialized using Skycoin's encoder:
	// outB1 := encoder.Serialize(int32(50))

	// Then we need to know where we can start writing in the CX stack. `out1Offset` will be, for example, 10313
	// which means that we can write the 4 bytes that represent the i32 integer 50 (which is the byte array [50 0 0 0])
	// to bytes 10313, 10314, 10315 and 10316.
	// out1Offset := GetFinalOffset(fp, out1)

	// And finally we write those bytes to the CX stack.
	// WriteMemory(out1Offset, outB1)
}

// op_http_HandleFunc lastly, let's see how we would associate functions to our router's endpoints.
func op_http_HandleFunc(prgrm *CXProgram) {
	expr := prgrm.GetExpr()
	fp := prgrm.GetFramePointer()

	// In this case, `http.HandleFunc` requires two inputs: the name of the endpoint (`inp1`)
	// and a function to be called when that endpoint is requested (`inp2`).
	inp1, inp2 := expr.Inputs[0], expr.Inputs[1]

	// Endpoint is a string, so we extract the value using `ReadStr`. There are other
	// helper functions for extracting stuff, such as `ReadBool`, `ReadI32`, etc. These are
	// defined in `cx/op.go`.
	endpoint := ReadStr(fp, inp1)

	// FOR NOW, extracting the function to be called needs to be done like this:
	fnName := inp2.Name
	pkgName := inp2.Package.Name
	// which simply gets the name of the variable. So the only type of function calls that are allowed for now
	// are like this: `http.HandleFunc("/foo", foo)`, where `foo` is a function defined in a package.
	// Something like `http.HandleFunc("/foo", bar.foo)`, where `bar` is a package, should be valid.

	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request){
		// We'll need to create another CX function that accepts input (a string, for example) and sends
		// it to `w.ResponseWriter`. Normally you'd do something like:
		// `fmt.Fprintf(w, "Hello, world!")`
		// but we'd need to make a CX structure that implements `http.ResponseWriter`.

		// `prgrm.ccallback` calls a CX function of name `fnName`, located at package `pkgName`
		// the `nil` in here should be a [][]byte argument that represents the ordered inputs to the function
		// to be called.
		prgrm.ccallback(expr, fnName, pkgName, nil)

		// As you can see, this part doesn't work as it is. You'd need to think on how to solve this.
		// A possible solution that comes to my mind is send to `w` the result of calling `ccallback`, but then
		// we'd need to modify how `ccallback` behaves or create an extension of it that actually writes something
		// to the stack (as you can see, everything called by `ccallback` is outputless).
	})

	// SOON we should have a helper function such as `ReadFunc(fp, inp1)` which returns
	// a pointer to a CXFunction (see `cx/cxfunction.go`) that we can use instead,
	// mainly to use unnamed function literals (lambda functions). This should be part
	// of an extension to my "functions as first-class objects" branch.

	// Finally, when you have defined the functions in this file, you need to add their OP codes to
	// cx/opcodes_base.go. Recompile CX, create a test.cx file where you import `http`, and test
	// your new functions.
}
