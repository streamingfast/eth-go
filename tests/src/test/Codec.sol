// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.10;

import "forge-std/Test.sol";
import "src/Codec.sol";

contract CodecTest is Test {
    Codec codec;

    function setUp() public {
        codec = new Codec();
    }

    function testFunFixedArraySubFixed() public {
        address[2] memory input;
        input[0] = 0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa;
        input[1] = 0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF;

        bytes memory actual = abi.encodeWithSignature(
            "funFixedArraySubFixed(address[2])",
            input
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"49508494000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa000000000000000000000000ffffffffffffffffffffffffffffffffffffffff"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunDynamicBoolArray() public {
        bool[] memory input = new bool[](2);
        input[0] = true;
        input[1] = false;

        bytes memory actual = abi.encodeWithSignature(
            "funDynamicBoolArray(bool[])",
            input
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"b0e615780000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunFixedArraySubDynamic() public {
        bytes memory bytes_0 = hex"aaaaaaaaaa";
        bytes memory bytes_1 = hex"ffffffffff";

        bytes[2] memory input;
        input[0] = bytes_0;
        input[1] = bytes_1;

        bytes memory actual = abi.encodeWithSignature(
            "funFixedArraySubDynamic(bytes[2])",
            input
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"cd9f57d00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005aaaaaaaaaa0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ffffffffff000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunBytes8Bytes16Bytes24Bytes32() public {
        bytes8 fixed_bytes_8_type = 0xc5abac1e99944b1d;
        bytes16 fixed_bytes_16_type = 0x57dbc30b9acfebfb86bcc5f9e2fe3fa0;
        bytes24 fixed_bytes_24_type = 0x04a81d8d5c3958b07e558ff8e58e1edf1871c14b34ecdc1c;
        bytes32 fixed_bytes_32_type = 0xf154bf9817019c089414b85e6c5a19fd5d1ea04c103fcd039314132b354ca184;

        bytes memory actual = abi.encodeWithSignature(
            "funBytes8Bytes16Bytes24Bytes32(bytes8,bytes16,bytes24,bytes32)",
            fixed_bytes_8_type,
            fixed_bytes_16_type,
            fixed_bytes_24_type,
            fixed_bytes_32_type
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"680657bdc5abac1e99944b1d00000000000000000000000000000000000000000000000057dbc30b9acfebfb86bcc5f9e2fe3fa00000000000000000000000000000000004a81d8d5c3958b07e558ff8e58e1edf1871c14b34ecdc1c0000000000000000f154bf9817019c089414b85e6c5a19fd5d1ea04c103fcd039314132b354ca184"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testMint() public {
        vm.recordLogs();

        (uint256 tokenId, uint256 nextId) = codec.mint();

        require(tokenId == 0, "Invalid tokenId result");
        require(nextId == 1, "Invalid nextId result");

        (uint256 tokenIdSecond, uint256 nextIdSecond) = codec.mint();

        require(tokenIdSecond == 1, "Invalid tokenId result");
        require(nextIdSecond == 2, "Invalid nextId result");

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 2, "Logs length invalid");

        assertEq(logs[0].topics.length, 2);
        assertEq(logs[0].topics[1], bytes32(uint256(0)));

        assertEq(logs[1].topics.length, 2);
        assertEq(logs[1].topics[1], bytes32(uint256(1)));
    }

    function testFunInt8() public {
        int8 value = -127;

        bytes memory actual = abi.encodeWithSignature("funInt8(int8)", value);

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"3036e687ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunInt32() public {
        int32 value = -898877731;

        bytes memory actual = abi.encodeWithSignature("funInt32(int32)", value);

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"d78caab3ffffffffffffffffffffffffffffffffffffffffffffffffffffffffca6c36dd"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunInt128Value0() public {
        int128 value = 0;

        bytes memory actual = abi.encodeWithSignature(
            "funInt128(int128)",
            value
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"5b3357ff0000000000000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunInt128ValueMinus1() public {
        int128 value = -1;

        bytes memory actual = abi.encodeWithSignature(
            "funInt128(int128)",
            value
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"5b3357ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunInt256() public {
        int256 value = -9809887317731;

        bytes memory actual = abi.encodeWithSignature(
            "funInt256(int256)",
            value
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"f70af73bfffffffffffffffffffffffffffffffffffffffffffffffffffff713f526b11d"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunInt8Int32Int64Int256() public {
        int8 value8 = -127;
        int32 value32 = -898877731;
        int64 value64 = -9809887317731;
        int256 value256 = -223372036854775808;

        bytes memory actual = abi.encodeWithSignature(
            "funInt8Int32Int64Int256(int8,int32,int64,int256)",
            value8,
            value32,
            value64,
            value256
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"db617e8fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81ffffffffffffffffffffffffffffffffffffffffffffffffffffffffca6c36ddfffffffffffffffffffffffffffffffffffffffffffffffffffff713f526b11dfffffffffffffffffffffffffffffffffffffffffffffffffce66c50e2840000"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunString() public {
        bytes memory actual = abi.encodeWithSignature(
            "funString(string)",
            "test"
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"b0d94419000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000047465737400000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testFunReturnsString() public {
        bytes memory actual = abi.encodeWithSignature("funReturnsString()");

        codec.logBytes(actual);

        require(
            bytesEquals(actual, hex"7a3719f0"),
            "Invalid input encode packed bytes"
        );

        (bool success, bytes memory output) = address(codec).call(actual);
        require(success, "call should have succeeed");

        codec.logBytes(output);

        require(
            bytesEquals(
                output,
                hex"000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000047465737400000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid output encode packed bytes"
        );
    }

    function testFunReturnsStringString() public {
        bytes memory actual = abi.encodeWithSignature(
            "funReturnsStringString()"
        );

        codec.logBytes(actual);

        require(
            bytesEquals(actual, hex"85032f7c"),
            "Invalid input encode packed bytes"
        );

        (bool success, bytes memory output) = address(codec).call(actual);
        require(success, "call should have succeeed");

        codec.logBytes(output);

        require(
            bytesEquals(
                output,
                hex"000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005746573743100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000057465737432000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid output encode packed bytes"
        );
    }

    function testFunFixedArrayAddressArrayUint256ReturnsUint256String() public {
        address[2] memory senders;
        senders[0] = 0xFffDB7377345371817F2b4dD490319755F5899eC;
        senders[1] = 0xFFFdb7377345371817F2B4DD490319755F5899EB;

        address[] memory receivers = new address[](3);
        receivers[0] = 0xaFfdb7377345371817f2b4Dd490319755f5899eC;
        receivers[1] = 0xbfFDB7377345371817F2b4dd490319755f5899eC;
        receivers[2] = 0xCffdb7377345371817F2b4Dd490319755F5899EC;

        bytes memory actual = abi.encodeWithSignature(
            "funFixedArrayAddressArrayUint256ReturnsUint256String(address[2],address[])",
            senders,
            receivers
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"b2e8fed2000000000000000000000000fffdb7377345371817f2b4dd490319755f5899ec000000000000000000000000fffdb7377345371817f2b4dd490319755f5899eb00000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000003000000000000000000000000affdb7377345371817f2b4dd490319755f5899ec000000000000000000000000bffdb7377345371817f2b4dd490319755f5899ec000000000000000000000000cffdb7377345371817f2b4dd490319755f5899ec"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, bytes memory output) = address(codec).call(actual);
        require(success, "call should have succeeed");

        codec.logBytes(output);

        require(
            bytesEquals(
                output,
                hex"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid output encode packed bytes"
        );
    }

    function testFunAll() public {
        address address_type = 0xFffDB7377345371817F2b4dD490319755F5899eC;
        bytes memory bytes_type = hex"b2";
        bytes8 fixed_bytes_8_type = 0xcf36ac4f97dc10d9;
        bytes32 fixed_bytes_32_type = 0xcf36ac4f97dc10d91fc2cbb20d718e94a8cbfe0f82eaedc6a4aa38946fb797cd;
        int256 fixed_int_256_type = -9809887317731;
        uint256 fixed_uint_256_type = 1827641804;
        bool bool_type = true;
        string memory string_type = "test";
        address[2] memory fixed_array_type;
        address[] memory array_type = new address[](0);

        bytes memory actual = abi.encodeWithSignature(
            "funAll(address,bytes,bytes8,bytes32,int256,uint256,bool,string,address[2],address[])",
            address_type,
            bytes_type,
            fixed_bytes_8_type,
            fixed_bytes_32_type,
            fixed_int_256_type,
            fixed_uint_256_type,
            bool_type,
            string_type,
            fixed_array_type,
            array_type
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"1af93c31000000000000000000000000fffdb7377345371817f2b4dd490319755f5899ec0000000000000000000000000000000000000000000000000000000000000160cf36ac4f97dc10d9000000000000000000000000000000000000000000000000cf36ac4f97dc10d91fc2cbb20d718e94a8cbfe0f82eaedc6a4aa38946fb797cdfffffffffffffffffffffffffffffffffffffffffffffffffffff713f526b11d000000000000000000000000000000000000000000000000000000006cef99cc000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000001a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001e00000000000000000000000000000000000000000000000000000000000000001b200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000474657374000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeed");
    }

    function testEmitEventIArrayAddress() public {
        vm.recordLogs();

        codec.emitEventIArrayAddress();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 2);
        assertEq(logs[0].topics[0], keccak256("EventIArrayAddress(address[])"));

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(logs[0].data, hex""),
            "Invalid input encode packed bytes"
        );
    }

    function testEmitEventUTuple1() public {
        vm.recordLogs();

        codec.emitEventUTuple1();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);

        // We keept both for reference documentation of the event ID
        assertEq(
            logs[0].topics[0],
            0xae8abdcc248ceb1263d6850d955eef93d14c59779f2aa0c12594bb7d8cc7f0e0
        );
        assertEq(logs[0].topics[0], keccak256("EventUTuple1((address))"));

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"000000000000000000000000db0de9288cf0713de91371969efcc9969dd94117"
            ),
            "Invalid input encode packed bytes"
        );
    }

    function testEmitEventUTupleBool() public {
        vm.recordLogs();

        codec.emitEventUTupleBool();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);

        // We keept both for reference documentation of the event ID
        assertEq(
            logs[0].topics[0],
            0xe46e0615228a85d593cefeae9bb5f9d1b6698858b635d549b40492afb258ff23
        );
        assertEq(logs[0].topics[0], keccak256("EventUTupleBool((bool))"));

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000001"
            ),
            "Invalid input encode packed bytes"
        );
    }

    function testEmitEventUArrayBool() public {
        vm.recordLogs();

        codec.emitEventUArrayBool();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);

        // We keept both for reference documentation of the event ID
        assertEq(
            logs[0].topics[0],
            0xee0cd0e55d575e4e32db712d239532b1104938ed2971f10d8b63e4aa4c17afb6
        );
        assertEq(logs[0].topics[0], keccak256("EventUArrayBool(bool[])"));

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );
    }

    function testEmitEventUFixedArrayString() public {
        vm.recordLogs();

        codec.emitEventUFixedArrayString();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);

        // We keept both for reference documentation of the event ID
        assertEq(
            logs[0].topics[0],
            0x2f66d1a00558d55ced0f61b550ca490f9718523b5181b89c06b24ed7752e137c
        );
        assertEq(
            logs[0].topics[0],
            keccak256("EventUFixedArrayString(string[2])")
        );

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000005666972737400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000067365636f6e640000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );
    }

    function testEmitEventUFixedArrayBool() public {
        vm.recordLogs();

        codec.emitEventUFixedArrayBool();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);

        // We keept both for reference documentation of the event ID
        assertEq(
            logs[0].topics[0],
            0x7105c6c17f0acf302d800b9cb7b8d17a738e2ed9fa7a4a02416262a25461be5d
        );
        assertEq(logs[0].topics[0], keccak256("EventUFixedArrayBool(bool[2])"));

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"
            ),
            "Invalid input encode packed bytes"
        );
    }

    // ===== Tests for tuples with dynamic components =====

    function testFunTupleWithBytes() public {
        // TupleWithBytes: (address signer, bytes metadata, uint64 value)
        // This tuple has a dynamic `bytes` field, so it should use offset-based encoding
        Codec.TupleWithBytes memory input = Codec.TupleWithBytes(
            0x1234567890123456789012345678901234567890,
            hex"deadbeef",
            42
        );

        bytes memory actual = abi.encodeWithSignature(
            "funTupleWithBytes((address,bytes,uint64))",
            input
        );

        codec.logBytes(actual);

        // Expected encoding:
        // - 4 bytes: method selector
        // - 32 bytes: offset to tuple data (0x20 = 32)
        // Then the tuple itself:
        // - 32 bytes: address (inline, padded)
        // - 32 bytes: offset to bytes data (0x60 = 96, relative to tuple start)
        // - 32 bytes: uint64 value (inline, padded)
        // - 32 bytes: bytes length (4)
        // - 32 bytes: bytes data (0xdeadbeef, padded)
        require(
            bytesEquals(
                actual,
                hex"1c09b0d7" // method selector for funTupleWithBytes((address,bytes,uint64))
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to tuple
                hex"0000000000000000000000001234567890123456789012345678901234567890" // address signer
                hex"0000000000000000000000000000000000000000000000000000000000000060" // offset to bytes
                hex"000000000000000000000000000000000000000000000000000000000000002a" // uint64 value = 42
                hex"0000000000000000000000000000000000000000000000000000000000000004" // bytes length = 4
                hex"deadbeef00000000000000000000000000000000000000000000000000000000" // bytes data
            ),
            "Invalid TupleWithBytes encoding"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeded");
    }

    function testFunTupleWithString() public {
        // TupleWithString: (uint256 id, string name)
        Codec.TupleWithString memory input = Codec.TupleWithString(123, "hello");

        bytes memory actual = abi.encodeWithSignature(
            "funTupleWithString((uint256,string))",
            input
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"087481b8" // method selector for funTupleWithString((uint256,string))
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to tuple
                hex"000000000000000000000000000000000000000000000000000000000000007b" // uint256 id = 123
                hex"0000000000000000000000000000000000000000000000000000000000000040" // offset to string (64 from tuple start)
                hex"0000000000000000000000000000000000000000000000000000000000000005" // string length = 5
                hex"68656c6c6f000000000000000000000000000000000000000000000000000000" // "hello"
            ),
            "Invalid TupleWithString encoding"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeded");
    }

    function testFunTupleWithMultipleDynamic() public {
        // TupleWithMultipleDynamic: (bytes data1, uint64 value, bytes data2)
        Codec.TupleWithMultipleDynamic memory input = Codec.TupleWithMultipleDynamic(
            hex"0102",
            100,
            hex"030405"
        );

        bytes memory actual = abi.encodeWithSignature(
            "funTupleWithMultipleDynamic((bytes,uint64,bytes))",
            input
        );

        codec.logBytes(actual);

        require(
            bytesEquals(
                actual,
                hex"2609836a" // method selector for funTupleWithMultipleDynamic((bytes,uint64,bytes))
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to tuple
                hex"0000000000000000000000000000000000000000000000000000000000000060" // offset to data1 (96 from tuple start)
                hex"0000000000000000000000000000000000000000000000000000000000000064" // uint64 value = 100
                hex"00000000000000000000000000000000000000000000000000000000000000a0" // offset to data2 (160 from tuple start)
                hex"0000000000000000000000000000000000000000000000000000000000000002" // data1 length = 2
                hex"0102000000000000000000000000000000000000000000000000000000000000" // data1 = 0x0102
                hex"0000000000000000000000000000000000000000000000000000000000000003" // data2 length = 3
                hex"0304050000000000000000000000000000000000000000000000000000000000" // data2 = 0x030405
            ),
            "Invalid TupleWithMultipleDynamic encoding"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeded");
    }

    function testEmitEventUTupleWithBytes() public {
        vm.recordLogs();

        codec.emitEventUTupleWithBytes();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);
        assertEq(
            logs[0].topics[0],
            keccak256("EventUTupleWithBytes((address,bytes,uint64))")
        );

        codec.logBytes(logs[0].data);

        // Event data encoding for tuple with dynamic component
        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to tuple
                hex"000000000000000000000000db0de9288cf0713de91371969efcc9969dd94117" // address
                hex"0000000000000000000000000000000000000000000000000000000000000060" // offset to bytes
                hex"000000000000000000000000000000000000000000000000000000000000002a" // uint64 = 42
                hex"0000000000000000000000000000000000000000000000000000000000000004" // bytes length = 4
                hex"deadbeef00000000000000000000000000000000000000000000000000000000" // bytes data
            ),
            "Invalid EventUTupleWithBytes data encoding"
        );
    }

    function testEmitEventUTupleWithString() public {
        vm.recordLogs();

        codec.emitEventUTupleWithString();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);
        assertEq(
            logs[0].topics[0],
            keccak256("EventUTupleWithString((uint256,string))")
        );

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to tuple
                hex"000000000000000000000000000000000000000000000000000000000000007b" // uint256 id = 123
                hex"0000000000000000000000000000000000000000000000000000000000000040" // offset to string
                hex"0000000000000000000000000000000000000000000000000000000000000005" // string length = 5
                hex"68656c6c6f000000000000000000000000000000000000000000000000000000" // "hello"
            ),
            "Invalid EventUTupleWithString data encoding"
        );
    }

    function testEmitEventUTupleWithMultipleDynamic() public {
        vm.recordLogs();

        codec.emitEventUTupleWithMultipleDynamic();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);
        assertEq(
            logs[0].topics[0],
            keccak256("EventUTupleWithMultipleDynamic((bytes,uint64,bytes))")
        );

        codec.logBytes(logs[0].data);

        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to tuple
                hex"0000000000000000000000000000000000000000000000000000000000000060" // offset to data1
                hex"0000000000000000000000000000000000000000000000000000000000000064" // uint64 = 100
                hex"00000000000000000000000000000000000000000000000000000000000000a0" // offset to data2
                hex"0000000000000000000000000000000000000000000000000000000000000002" // data1 length = 2
                hex"0102000000000000000000000000000000000000000000000000000000000000" // data1
                hex"0000000000000000000000000000000000000000000000000000000000000003" // data2 length = 3
                hex"0304050000000000000000000000000000000000000000000000000000000000" // data2
            ),
            "Invalid EventUTupleWithMultipleDynamic data encoding"
        );
    }

    function testFunTupleWithNestedDynamic() public {
        // TupleWithNestedDynamic: (address owner, InnerDynamic inner, uint64 timestamp)
        // where InnerDynamic: (uint256 id, string description)
        Codec.InnerDynamic memory inner = Codec.InnerDynamic(456, "nested");
        Codec.TupleWithNestedDynamic memory input = Codec.TupleWithNestedDynamic(
            0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa,
            inner,
            789
        );

        bytes memory actual = abi.encodeWithSignature(
            "funTupleWithNestedDynamic((address,(uint256,string),uint64))",
            input
        );

        codec.logBytes(actual);

        // Expected encoding:
        // - 4 bytes: method selector
        // - 32 bytes: offset to outer tuple data (0x20)
        // Then outer tuple:
        // - 32 bytes: address owner (inline)
        // - 32 bytes: offset to inner tuple (0x60 = 96)
        // - 32 bytes: uint64 timestamp (inline)
        // Then inner tuple at offset 0x60:
        // - 32 bytes: uint256 id (inline)
        // - 32 bytes: offset to string (0x40 = 64 relative to inner tuple start)
        // - 32 bytes: string length (6)
        // - 32 bytes: string data ("nested")
        require(
            bytesEquals(
                actual,
                hex"56e67bfe" // method selector for funTupleWithNestedDynamic((address,(uint256,string),uint64))
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to outer tuple
                hex"000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // address owner
                hex"0000000000000000000000000000000000000000000000000000000000000060" // offset to inner tuple (96 from outer tuple start)
                hex"0000000000000000000000000000000000000000000000000000000000000315" // uint64 timestamp = 789
                hex"00000000000000000000000000000000000000000000000000000000000001c8" // uint256 id = 456
                hex"0000000000000000000000000000000000000000000000000000000000000040" // offset to string (64 from inner tuple start)
                hex"0000000000000000000000000000000000000000000000000000000000000006" // string length = 6
                hex"6e65737465640000000000000000000000000000000000000000000000000000" // "nested"
            ),
            "Invalid TupleWithNestedDynamic encoding"
        );

        (bool success, ) = address(codec).call(actual);
        require(success, "call should have succeeded");
    }

    function testEmitEventUTupleWithNestedDynamic() public {
        vm.recordLogs();

        codec.emitEventUTupleWithNestedDynamic();

        Vm.Log[] memory logs = vm.getRecordedLogs();
        require(logs.length == 1, "Logs length invalid");

        assertEq(logs[0].topics.length, 1);
        assertEq(
            logs[0].topics[0],
            keccak256("EventUTupleWithNestedDynamic((address,(uint256,string),uint64))")
        );

        codec.logBytes(logs[0].data);

        // Event data encoding for nested tuple with dynamic component
        require(
            bytesEquals(
                logs[0].data,
                hex"0000000000000000000000000000000000000000000000000000000000000020" // offset to outer tuple
                hex"000000000000000000000000aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // address owner
                hex"0000000000000000000000000000000000000000000000000000000000000060" // offset to inner tuple
                hex"0000000000000000000000000000000000000000000000000000000000000315" // uint64 timestamp = 789
                hex"00000000000000000000000000000000000000000000000000000000000001c8" // uint256 id = 456
                hex"0000000000000000000000000000000000000000000000000000000000000040" // offset to string
                hex"0000000000000000000000000000000000000000000000000000000000000006" // string length = 6
                hex"6e65737465640000000000000000000000000000000000000000000000000000" // "nested"
            ),
            "Invalid EventUTupleWithNestedDynamic data encoding"
        );
    }

    // ===== End of tests for tuples with dynamic components =====

    // FIXME: Shared for all tests ..., copied from test/PersonalSigning
    // Compares the 'len' bytes starting at address 'addr' in memory with the 'len'
    // bytes starting at 'addr2'.
    // Returns 'true' if the bytes are the same, otherwise 'false'.
    function memoryEquals(
        uint256 addr,
        uint256 addr2,
        uint256 len
    ) internal pure returns (bool equal) {
        assembly {
            equal := eq(keccak256(addr, len), keccak256(addr2, len))
        }
    }

    // Checks if two `bytes memory` variables are equal. This is done using hashing,
    // which is much more gas efficient then comparing each byte individually.
    // Equality means that:
    //  - 'self.length == other.length'
    //  - For 'n' in '[0, self.length)', 'self[n] == other[n]'
    function bytesEquals(
        bytes memory self,
        bytes memory other
    ) internal pure returns (bool equal) {
        if (self.length != other.length) {
            return false;
        }
        uint256 addr;
        uint256 addr2;

        assembly {
            addr := add(
                self,
                /*BYTES_HEADER_SIZE*/
                32
            )
            addr2 := add(
                other,
                /*BYTES_HEADER_SIZE*/
                32
            )
        }

        equal = memoryEquals(addr, addr2, self.length);
    }
}
