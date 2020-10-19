# Proposal: Zeroization of Secrets after usage

The goal is to minimize the lifetime of a Secret (and it's footprint) in RAM. 
In order to achieve this goal, Secretless must:

1. Be able to zeroize the original Secret and any deratives thereof. This is from the standpoint of technical feasibility of zeroing data types.
2. Once a Secret has been used for it's intended purpose, carry out the zeroization.

Ideally, 2 would happen in a centralized fashion to avoid the need for repeated boilerplate at the site of each derivative.


## Journey of Secrets
It is worth noting the journey of Secrets in Secretless.

```
Resolver.#Resolve -> Provider.#GetValues |-> Handler -> Listener -> Target Backend
                                        |
					|-> EventNotifier.#ResolveSecret
```
## Considerations

Below are some of the considerations made in coming together with this proposal. Some of these need to be validated.

1. Secrets arrive from the wire as byte slices. This is the only copy of the Secret in memory. If this is not the case then we'll need to think of ways to address this separately.
2. Byte slices are passed by reference from place to place, unless resized or concatenated. Avoid uncontrolled resizes at all cost - we don't know what happens to the memory of the byte slice on resize. If we need to concatenate byte slices we can use the following utility function `ConcatBytes` and recognize the result as a derivative:
```
func ConcatBytes(bss... []byte) []byte {
	size := 0
	for _, bs := range bss  {
		size += len(bs)
	}

	newByteSlice := make([]byte, size)

	// avoided using append because every time the byte slice is resized
	// memory will be copied over to the new byte slice
	// and i guess the old byte slice will be GC'd - which isn't good enough
	i := 0
	for _, bs := range bss  {
		for _, byte := range bs {
			newByteSlice[i] = byte
			i += 1
		}
	}
	return newByteSlice
}
```
3. We need to mindful of the passage of the Secret deravitives through the libraries responsible for using the Secret derivatives to establish a connection. e.g. Does the http listener create additional Secret deravitives, or is it enough to clean up the ones we pass to it?
4. Strings in Go are by default immutable. An attempt to modify a string through reflection or otherwise will result in PANIC! It's possible to mutate them using CGO but that's it's own can of worms. Immutable strings leave us at the whims of the GC. Mutable Strings are possible by having a byte slice (which is mutable) be coerced into looking like a String. See below:
```
func ByteBoundString(b []byte) string {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bytesHeader := &reflect.StringHeader{
		Data: header.Data,
		Len: header.Len,
	}
	return *(*string)(unsafe.Pointer(bytesHeader))
}
```
5. Given all the above, all Secret derivatives are byte slices that we can keep track of.
6. Byte slices can zeroized by running the following utility function:
```
func ZeroizeByteSlice(bs []byte) {
	for byteIndex := range bs {
		bs[byteIndex] = 0
	}
}
```

## Proposal

Given the above considerations, it is possible to meet requirement 1. of our goal. To reiterate, all Secret derivatives are either byte slices or strings. The strings can be made from byte slices so all Secret derivatives are byte slices. Byte slices are zeroizable.

In order to meet requirement 2., Secretless needs a mechanism of:
1. Tracking all the Secret derivatives
2. Broadcast when the Secret has been used for it's inteded purpose

The recommendation per Secret usage session is as follows:

- [ ] Use `context.Context` to store a registry for all the Secret derivatives via `context.WithValue`. 
- [ ] Access to the registry needs to be thread/goroutine safe - perhaps use mutex.
- [ ] Thread context through the **journey of Secrets**. This will require the modification of some of our interfaces. NOTE: this is something that needs to happen anyway, the handler interface is currently a hodge podge of no at all universal functions
- [ ] Once the Secret has been used:
  + Use `context.WithCancel` and call `cancel` so that interested party can carry out an clean up logic
  + Zeroize collection of Secret derivatives stored in context.

## Further consideration

1. Audit `Listener -> Target Backend`
1. Audit `Provider.#GetValues`
1. Reconsider `EventNotifier.#ResolveSecret`, why do we need this ?
