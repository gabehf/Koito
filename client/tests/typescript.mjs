import ts from "typescript";

// Use the stable compiler already installed for type checking, without Node's
// experimental stripTypeScriptTypes API or warning suppression.
export function compileTypeScript(source) {
  const result = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.ESNext,
      target: ts.ScriptTarget.ES2022,
    },
    reportDiagnostics: true,
  });
  if (result.diagnostics?.length) {
    throw new Error(
      ts.formatDiagnosticsWithColorAndContext(result.diagnostics, {
        getCanonicalFileName: (name) => name,
        getCurrentDirectory: () => process.cwd(),
        getNewLine: () => "\n",
      }),
    );
  }
  return result.outputText;
}
