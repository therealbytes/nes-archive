// ==== CONSTANTS ====

NES_WIDTH = 256;
NEW_HEIGHT = 240;

// ==== KEYBOARD INPUT ====

// Create an object to store the key states
const keyStates = {};
const pendingKeyStates = {};

// Event listener for keydown event
window.addEventListener('keydown', (event) => {
    // Set the corresponding key state to true when a key is pressed
    keyStates[event.code] = true;
});

// Event listener for keyup event
window.addEventListener('keyup', (event) => {
    // Set the corresponding key state to false when a key is released
    pendingKeyStates[event.code] = false;
});

function updateKeyStates() {
    for (const [key, value] of Object.entries(pendingKeyStates)) {
        keyStates[key] = value;
    }
}

// Function to check if a key is currently pressed
function isKeyPressed(keyCode) {
    return keyStates[keyCode] === true;
}

function isKeyPressedUint8(keyCode) {
    return isKeyPressed(keyCode) ? 1 : 0;
}

// ==== WEBASSEMBLY ====

if (WebAssembly) {
    // WebAssembly.instantiateStreaming is not currently available in Safari
    if (WebAssembly && !WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }

    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then(async (result) => {
        go.run(result.instance);

        // Call the setup function
        const staticData = new Uint8Array([1, 2, 3, 4]);
        const dynamicData = new Uint8Array([5, 6, 7, 8]);

        setup(staticData, dynamicData);

        // Get the canvas element
        const canvas = document.getElementById('canvas');
        const context = canvas.getContext('2d');

        // Call the step function and render the result
        function renderStep(buttons) {
            const imageData = context.createImageData(NES_WIDTH, NEW_HEIGHT);
            const rawImageData = new Uint8Array(NES_WIDTH * NEW_HEIGHT * 4);
            let startTime = performance.now();
            step(buttons, rawImageData);
            let endTime = performance.now();
            console.log("Tick took " + (endTime - startTime) + " milliseconds.");
            startTime = performance.now();
            imageData.data.set(rawImageData);
            context.putImageData(imageData, 0, 0);
            endTime = performance.now();
            console.log("Render took " + (endTime - startTime) + " milliseconds.");
        }

        // Start the rendering
        while (true) {
            const buttons = new Uint8Array([
                isKeyPressedUint8("KeyZ"),
                isKeyPressedUint8("KeyX"),
                isKeyPressedUint8("ShiftRight"),
                isKeyPressedUint8("Enter"),
                isKeyPressedUint8("ArrowUp"),
                isKeyPressedUint8("ArrowDown"),
                isKeyPressedUint8("ArrowLeft"),
                isKeyPressedUint8("ArrowRight"),
            ]);
            console.log({
                "ButtonA" : buttons[0],
                "ButtonB" : buttons[1],
                "ButtonSelect" : buttons[2],
                "ButtonStart" : buttons[3],
                "ButtonUp" : buttons[4],
                "ButtonDown" : buttons[5],
                "ButtonLeft" : buttons[6],
                "ButtonRight" : buttons[7],
            });
            renderStep(buttons);
            updateKeyStates();
            await new Promise(r => setTimeout(r, 1000));
        }
    });
} else {
    console.log("WebAssembly is not supported in your browser")
}
