if (WebAssembly) {
    if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }
    const go = new Go();
    let mod, inst;
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
        console.log("WebAssembly module loaded");
        mod = result.module;
        inst = result.instance;
        go.run(inst)
        inst = WebAssembly.instantiate(mod, go.importObject); // reset instance
    }).catch((err) => {
        console.error(err);
    });
} else {
    console.log("WebAssembly is not supported in your browser")
}
