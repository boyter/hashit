package processor

import (
	"reflect"
	"testing"
)

func Test_parseHashdeepFile(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []auditRecord
	}{
		{
			name: "two entries",
			args: sampleHashdeepAudit,
			want: []auditRecord{
				{
					Size:     "1067",
					MD5:      "227f999ca03b135a1b4d69bde84afb16",
					SHA256:   "fb3f44f5e74b957107f89b027896250ecff74718b1fa8bf0566874e142e54351",
					Filename: "/Users/boyter/Documents/projects/hashit/LICENSE",
				},
				{
					Size:     "1051",
					MD5:      "67bd18902fe2faa001d9eaa1d36a44ae",
					SHA256:   "4812954b042ad3e1c856766ea385a25c521fbfe416002ab767b8e6e4c256aaf9",
					Filename: "/Users/boyter/Documents/projects/hashit/go.mod",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := parseHashdeepFile(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseHashdeepFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var sampleHashdeepAudit = `%%%% HASHDEEP-1.0
%%%% size,md5,sha256,filename
## Invoked from: /Users/boyter/Documents/projects
## $ hashdeep -r hashit -c md5,sha1,sha256
## 
1067,227f999ca03b135a1b4d69bde84afb16,fb3f44f5e74b957107f89b027896250ecff74718b1fa8bf0566874e142e54351,/Users/boyter/Documents/projects/hashit/LICENSE
1051,67bd18902fe2faa001d9eaa1d36a44ae,4812954b042ad3e1c856766ea385a25c521fbfe416002ab767b8e6e4c256aaf9,/Users/boyter/Documents/projects/hashit/go.mod`
