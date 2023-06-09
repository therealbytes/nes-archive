const staticHash = "0xda2437bb81b1a07d5e2832768ba41f1a43cf060ba5a2db3ac0265361220ed82c"
const dynHash = "0x4123f2d81428f7090218f975b941122f3797aeb8f97bf7d1ef6e87491c920a5c"

function hexToUint8Array(hexString) {
    if (hexString.length % 2 !== 0) {
        console.error("Invalid hexString");
        return;
    }
    if (hexString.startsWith("0x")) {
        hexString = hexString.slice(2);
    }
    var bytes = new Uint8Array(hexString.length / 2);
    for (var i = 0; i < hexString.length; i += 2) {
        bytes[i / 2] = parseInt(hexString.substr(i, 2), 16);
    }
    return bytes;
}

async function fetchPreimage(hashStr) {
    const resp = await fetch(`/preimages/${hashStr}`);
    const buffer = await resp.arrayBuffer();
    const preimage = new Uint8Array(buffer);
    return preimage;
}

const staticHashBytes = hexToUint8Array(staticHash);
const dynHashBytes = hexToUint8Array(dynHash);

if (WebAssembly) {
    if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then(async (result) => {
        console.log("WebAssembly module loaded");
        go.run(result.instance);
        const api = window.NesAPI();
        console.log("NesAPI loaded", api);

        api.setPreimage(staticHashBytes, await fetchPreimage(staticHash));
        api.setPreimage(dynHashBytes, await fetchPreimage(dynHash));

        api.start();
        api.setCartridge(staticHashBytes, dynHashBytes);

        setInterval(() => {
            const activity = api.getActivity();
            const jsonString = new TextDecoder().decode(activity);
            const jsonObject = JSON.parse(jsonString);
            console.log("Activity", jsonObject);
        }, 10000);
    }).catch((err) => {
        console.error(err);
    });
} else {
    console.log("WebAssembly is not supported in your browser")
}
