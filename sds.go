/**
 * MIT License
 *
 * <p>Copyright (c) 2020 mixmicro
 *
 * <p>Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the "Software"), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute,
 * sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * <p>The above copyright notice and this permission notice shall be included in all copies or
 * substantial portions of the Software.
 *
 * <p>THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
 * BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */
package redis_structure

const (
	SDS_TYPE_5 byte = iota + 1
	SDS_TYPE_8
	SDS_TYPE_16
	SDS_TYPE_32
	SDS_TYPE_64
)

const (
	SDS_5  = 1 << 5
	SDS_8  = 1 << 8
	SDS_16 = 1 << 16
	SDS_32 = 1 << 32
	SDS_64 = 1 << 64
)

type sds struct {
	flags byte
	buf   []rune
}

type sdshdr5 struct {
	sds
}

type sdshdr8 struct {
	sds
	len   uint8
	alloc uint8
}

type sdshdr16 struct {
	sds
	len   uint16
	alloc uint16
}

type sdshdr32 struct {
	sds
	len   uint32
	alloc uint32
}

type sdshdr64 struct {
	sds
	len   uint64
	alloc uint64
}

// Create a empty sds.
func NewSdsEmpty() *sds {
	return NewSds("", 0)
}

// Create a new sds string with the content specified by the (val and size).
func NewSds(val string, initlen int) *sds {
	sdsType := sdsReqType(initlen)

	// Empty strings are usually created in ordered to append. Use type 8 since type 5 is not good at this.
	if sdsType == SDS_TYPE_5 && initlen == 0 {
		sdsType = SDS_TYPE_8
	}
	return &sds{
		flags: 0,
		buf:   nil,
	}
}

// Free an sds string. No operation is performed if 's' is Null.
func (s *sds) FreeSds() {

}

// Modify an sds string in-place to make it empty(zero length).
// However all the existing buffer is not discarded but set as free space
// so that next append operations will not require allocations up to number of bytes previously available.
func (s *sds) ClearSds() {

}

// Append the specified sds to the existing sds.
func (s *sds) Addsds(val string) {

}

func (s *sds) sdscatlen(val string, len int) {

}

func (s *sds) sdsMakeRoomFor(len int) {

}

// Determine which type to used via initlen
func sdsReqType(initlen int) byte {
	if initlen < SDS_5 {
		return SDS_TYPE_5
	}

	if initlen < SDS_8 {
		return SDS_TYPE_8
	}

	if initlen < SDS_16 {
		return SDS_TYPE_16
	}

	if initlen < SDS_32 {
		return SDS_TYPE_32
	}

	if initlen < SDS_64 {
		return SDS_TYPE_64
	}

	return SDS_TYPE_32
}
