# magicbytes

package magicbytes // import "github.com/irukeru/binalyze-go-coding-challange/pkg/magicbytes"


CONSTANTS

const MaxMetaArrayLength = 1000
    MaxMetaArrayLength holds Maximum length for meta type array


VARIABLES

var ErrMetaArrayLengthExceeded = errors.New("Meta array length exceeded max value")
    ErrMetaArrayLengthExceeded holds meta array length exceeded max value error


FUNCTIONS

func Search(ctx context.Context, targetDir string, metas []*Meta, onMatch OnMatchFunc) error
    Search searches the given target directory to find files recursively using
    meta information. For every match, onMatch callback is called concurrently.


TYPES

type Meta struct {
        Type   string // name of the file/meta type.
        Bytes  []byte // magical bytes.
        Offset int64  // offset of the magical bytes from the file start position.
}
    Meta holds the name, magical bytes, and offset of the magical bytes to be
    searched.

type OnMatchFunc func(path, metaType string) bool
    OnMatchFunc represents a function to be called when Search function finds a
    match. Returning false must immediately stop Search process.