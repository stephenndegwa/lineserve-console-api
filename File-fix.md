Fix all the linting and syntax errors in the internal\services\subnet.go file. Use idiomatic Go practices and align the implementation with the Gophercloud v2 SDK.

🔍 Reference: A local folder named gophercloud is attached — it contains the official cloned GitHub repository for Gophercloud. Refer to this for correct usage patterns, especially with regards to:

Authentication

Compute service usage

Pagination and client initialization

📌 Important Context:

We are using Gophercloud v2 (not legacy v1).

The current stable version is v2.7.0.

You must use the new import path:

go
Copy
Edit
import "github.com/gophercloud/gophercloud/v2"
Run go mod tidy after adjusting imports, if necessary.

💡 Expectations:

Fix any linting issues (golangci-lint, go vet, etc.).

Ensure all method calls, structs, and interfaces match what’s defined in Gophercloud v2.

Output should be a corrected version of internal\services\subnet.go.