# build.sh
#!/bin/bash
solc --abi --bin \
  contracts/MyToken.sol \
  --base-path . \
  --include-path node_modules \
  -o output
