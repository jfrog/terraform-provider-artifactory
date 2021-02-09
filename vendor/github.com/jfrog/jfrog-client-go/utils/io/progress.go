package io

import "io"

// You may implement this interface to display progress indication of files transfer (upload / download)
type Progress interface {
	// Initializes a new progress indication for a new file transfer.
	// Input: 'total' - file size, 'prefix' - optional description, 'filePath' - path of the file being transferred (for description purposes only).
	// Output: progress indication id
	New(total int64, prefix, filePath string) (id int)
	// Replaces an indication (with the 'replaceId') when completed. Used when an additional work is done as part of the transfer.
	NewReplacement(replaceId int, prefix, filePath string) (id int)
	// Used to wrap an io.Reader in order to track the bytes reading count of the file transfer, and update indication 'id' accordingly.
	ReadWithProgress(id int, reader io.Reader) io.Reader
	// Aborts a progress indication. Called on both successful and unsuccessful operations
	Abort(id int)
	// Quits the whole progress mechanism
	Quit()
}
