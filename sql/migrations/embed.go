// The migrations package exists solely to embed the .sql migrations so the application can access them.
package migrations

import "embed"

//go:embed *
var FS embed.FS
