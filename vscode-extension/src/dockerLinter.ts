import * as vscode from 'vscode';
import * as path from 'path';
import * as child_process from 'child_process';
import {ThrottledDelayer} from './async';


const DOCKER_LANGUAGE = 'dockerfile';
export default class ShellCheckProvider implements vscode.CodeActionProvider {
    private readonly diagnosticCollection: vscode.DiagnosticCollection;

    constructor(private readonly context: vscode.ExtensionContext) {
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection();

        let disposable = vscode.commands.registerCommand('docker-lint.runLint', () => {
            vscode.window.showInformationMessage('Hello World from docker-lint!');
        });
        context.subscriptions.push(disposable);
        vscode.workspace.onDidOpenTextDocument(this.triggerLint, this, context.subscriptions);
        vscode.workspace.textDocuments.forEach(this.triggerLint, this);
    }

    provideCodeActions(document: vscode.TextDocument, range: vscode.Range | vscode.Selection, context: vscode.CodeActionContext, token: vscode.CancellationToken): vscode.ProviderResult<(vscode.CodeAction | vscode.Command)[]> {
        // Would be nice to have some examples where linter can provide quick fixes
        const actions: vscode.CodeAction[] = [];
        return actions;
    }

    private triggerLint(doc: vscode.TextDocument): void {
        if (doc.languageId !== DOCKER_LANGUAGE || doc.uri.scheme === 'git') {
            return;
        }

        let delayer = new ThrottledDelayer<void>(250);
        delayer.trigger(() => this.runLint(doc));
    }

    private runLint(dockerfile: vscode.TextDocument): Promise<void> {
        return new Promise<void>(( resolve, reject) => {
            let args = ['lint', dockerfile.fileName];
            const options = { cwd: path.dirname(dockerfile.fileName) } ;
            const childProcess =child_process.spawn('docker', args, options);
            childProcess.on('error', (error: NodeJS.ErrnoException) => {
                let message = `Failed to run docker lint: [${error.code}] ${error.message}`;
                vscode.window.showErrorMessage(message);
                resolve();
            });

            if (childProcess.pid) {
                childProcess.stdout.setEncoding('utf-8');
                const output: string[] = [];
                childProcess.stdout
                    .on('data', (data: Buffer) => {
                        output.push(data.toString());
                    })
                    .on('end', () => {
                        this.diagnosticCollection.set(dockerfile.uri, parseDiagnostics(output.join('').trim()));
                        resolve();
                    });
                    childProcess.stderr.on('data', (data: Buffer) => {
                        output.push(data.toString());
                    })
            } else {
                resolve();
            }
        });
    }

    public dispose(): void {
    }
}

function parseDiagnostics(textOutput: string) {
    let lines = textOutput.split("\n");
    let diags: vscode.Diagnostic[] = [];
    lines.forEach((line) => {
        let fields = line.split(' ');
        let pos = fields[0];
        let id = fields[1];
        let message = line.replace(pos + " " + id + " ", "");
        let posItems = pos.split(':');
        let lineNumber = parseInt(posItems[posItems.length - 1]);
        let start = new vscode.Position(lineNumber - 1, 0);
        let end = new vscode.Position(lineNumber, 0);
        let range = new vscode.Range(start, end);
        diags.push(new vscode.Diagnostic(range, message, vscode.DiagnosticSeverity.Warning));
    });
    return diags;
}
