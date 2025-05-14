# Usage

```go
import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Output struct {
	Header   interface{} `json:"header"`
	Proof    []string    `json:"proof"`
	TxnIndex uint64      `json:"txnIndex"`
	TxnCount uint64 	 `json:"txnCount"`
}

func main() {
	cmd := exec.Command("./verifytx", "-rpc", "http://localhost:9545", "-tx", "0x5e0262fa74c7c9dc436fedc5f414f7d7ed71bdd54c7cc305a20c09b450783220")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var parsed Output
	if err := json.Unmarshal(out, &parsed); err != nil {
		panic(err)
	}

	fmt.Printf("Txn Index: %d\n", parsed.TxnIndex)
}
```
