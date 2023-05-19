if (WebAssembly) {
    // WebAssembly.instantiateStreaming is not currently available in Safari
    if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }

    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
        go.run(result.instance);

        // Call the setup function
        const staticData = new Uint8Array([1, 2, 3, 4]);
        const dynamicData = new Uint8Array([5, 6, 7, 8]);

        setup(staticData, dynamicData);

        // Get the canvas element
        const canvas = document.getElementById('canvas');
        const context = canvas.getContext('2d');

        // Call the step function and render the result
        function renderStep() {
            const actionData = new Uint8Array();
            const rawImageData = new Uint8Array(canvas.width * canvas.height * 4);
            step(actionData, rawImageData);
            console.log(rawImageData);
            
            const imageData = context.createImageData(canvas.width, canvas.height);

            // Copy the image bytes to the ImageData
            imageData.data.set(rawImageData);

            // Render the ImageData on the canvas
            context.putImageData(imageData, 0, 0);

            // Request the next animation frame for continuous rendering
            // requestAnimationFrame(renderStep);
        }

        // Start the rendering
        renderStep();
    });
} else {
    console.log("WebAssembly is not supported in your browser")
}
