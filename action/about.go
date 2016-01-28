package action

import "github.com/Masterminds/glide/msg"

const aboutMessage = `
Glide: Vendor Package Management for Go. Manage your vendor and vendored
packages with ease.

Name:
    Aside from being catchy, "glide" is a contraction of "Go Elide". The idea is
    to compress the tasks that normally take us lots of time into a just a few
    seconds.

To file issues, obtain the source, or learn more visit:
    https://github.com/Masterminds/glide

Glide is licensed under the MIT License:

    Copyright (C) 2014-2015, Matt Butcher and Matt Farina
    Copyright (C) 2015, Google

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in
    all copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
    THE SOFTWARE.`

// About prints information about Glide.
func About() {
	msg.Puts(aboutMessage)
}
