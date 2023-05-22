function hexToUint8Array(hexString) {
    if (hexString.length % 2 !== 0) {
        console.error("Invalid hexString");
        return;
    }
    var bytes = new Uint8Array(hexString.length / 2);
    for (var i = 0; i < hexString.length; i += 2) {
        bytes[i / 2] = parseInt(hexString.substr(i, 2), 16);
    }
    return bytes;
}

const staticHash = "0xebefff5d04586f1d5ba0d052d1a06f2535c5dd92be22c289295442b1048fe872"
const staticHashBytes = hexToUint8Array(staticHash.slice(2));

const dynHash = "0x4123f2d81428f7090218f975b941122f3797aeb8f97bf7d1ef6e87491c920a5c"
const dynHashBytes = hexToUint8Array(dynHash.slice(2));

if (WebAssembly) {
    if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
        console.log("WebAssembly module loaded");
        go.run(result.instance);
        const api = window.NesAPI();
        console.log("NesAPI loaded", api);
        api.start();
        api.setCartridge(staticHashBytes, dynHashBytes);
    }).catch((err) => {
        console.error(err);
    });
} else {
    console.log("WebAssembly is not supported in your browser")
}
