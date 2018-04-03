# Copyright 2017 Intel Corporation.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

PATH_TO_MK = ../../mk
IMAGENAME =
EXECUTABLES =

bngdict.go: dictionary.bng
	radius-dict-gen -package radius -output bngdict.go dictionary.bng

testing: bngdict.go
	go test

include $(PATH_TO_MK)/leaf.mk
