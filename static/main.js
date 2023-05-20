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
    }).catch((err) => {
        console.error(err);
    });
} else {
    console.log("WebAssembly is not supported in your browser")
}
