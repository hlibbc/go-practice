const path = require("path");
const fs = require("fs");
const solc = require("solc");

const contractPath = path.resolve(__dirname, "contracts", "MyToken.sol");
const source = fs.readFileSync(contractPath, "utf8");

function findImports(importPath) {
  if (importPath.startsWith("@openzeppelin/")) {
    const ozPath = path.resolve(__dirname, "node_modules", importPath);
    return { contents: fs.readFileSync(ozPath, "utf8") };
  } else {
    return { error: "File not found" };
  }
}

const input = {
  language: "Solidity",
  sources: {
    "MyToken.sol": {
      content: source,
    },
  },
  settings: {
    outputSelection: {
      "*": {
        "*": ["abi", "evm.bytecode"],
      },
    },
  },
};

const output = JSON.parse(solc.compile(JSON.stringify(input), { import: findImports }));

if (output.errors) {
  let hasError = false;
  for (let error of output.errors) {
    console.error(error.formattedMessage);
    if (error.severity === "error") hasError = true;
  }
  if (hasError) {
    console.error("컴파일 중 오류가 발생했습니다.");
    process.exit(1);
  }
}

const abi = output.contracts["MyToken.sol"]["MyToken"].abi;
const bytecode = output.contracts["MyToken.sol"]["MyToken"].evm.bytecode.object;

fs.mkdirSync("./build", { recursive: true });

fs.writeFileSync("./build/MyToken.abi", JSON.stringify(abi, null, 2));
fs.writeFileSync("./build/MyToken.bin", bytecode);

console.log("ABI와 BIN 파일이 성공적으로 생성되었습니다.");
