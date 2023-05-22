KEY=0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6
RPC=http://localhost:9545
ADDRESS=0x0000000000000000000000000000000000000080

for file in ./preimages/*
do
    PREIMAGE=$(xxd -p ./preimages/0x4123f2d81428f7090218f975b941122f3797aeb8f97bf7d1ef6e87491c920a5c.bin  | tr -d '\n')
    HASH=$(basename $file)
    HASH=${HASH%.*}
    cast send $ADDRESS "addPreimage(bytes memory) returns (bytes32)" "$PREIMAGE" --private-key $KEY --rpc-url $RPC
done
