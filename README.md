# Verilog Language Server Extension

Inspired by the lack of editor support in existing verilog extensions for VSCode, this repository contains a powerful, hand-written Verilog language server and an extension to use it with VSCode.

## Functionality

This Language Server works for Verilog files.

Current Features:

- Completions
- Go To Definition
- Token Highlighting

Roadmap Features:

- Parsing Module Insides
- Error-tolerant parser
- Improved Completions
- Diagnostics

## Running the Sample

- Run `npm install` in this folder. This installs all necessary npm modules in both the client and server folder
- Open VS Code on this folder.
- Press Ctrl+Shift+B to start compiling the client and server in [watch mode](https://code.visualstudio.com/docs/editor/tasks#:~:text=The%20first%20entry%20executes,the%20HelloWorld.js%20file.).
- Switch to the Run and Debug View in the Sidebar (Ctrl+Shift+D).
- Select `Launch Client` from the drop down (if it is not already).
- Press ▷ to run the launch config (F5).
- In the [Extension Development Host](https://code.visualstudio.com/api/get-started/your-first-extension#:~:text=Then%2C%20inside%20the%20editor%2C%20press%20F5.%20This%20will%20compile%20and%20run%20the%20extension%20in%20a%20new%20Extension%20Development%20Host%20window.) instance of VSCode, open a Verilog Document.
  - You should now be able to use the extension and see the autocomplete functionality
