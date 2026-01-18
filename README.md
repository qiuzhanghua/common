# Common Utilities for golang

## Usage

### tar.gz

```go
import "github.com/qiuzhanghua/common/tgz"
import "github.com/qiuzhanghua/common/tz"
import "github.com/qiuzhanghua/common/util"
```

can be used as submodule with git

### Huggingface

```go
import "github.com/qiuzhanghua/common/hf"

pathOfModel, _ := hf.HfModelPath("intfloat/e5-mistral-7b-instruct")

```

### Add zst support

add github.com/klauspost/compress, so go 1.23 required
(zstd will be supported natively after go 1.27)
