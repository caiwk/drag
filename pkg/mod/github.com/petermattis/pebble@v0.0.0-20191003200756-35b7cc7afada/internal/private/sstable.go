// Copyright 2019 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package private

// SSTableWriterDisableKeyOrderChecks is a hook for disabling the key ordering
// invariant check performed by sstable.Writer. It is intended for internal use
// only in the construction of invalid sstables for testing. See
// tool/make_test_sstables.go.
var SSTableWriterDisableKeyOrderChecks func(interface{})
