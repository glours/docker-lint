import * as vscode from 'vscode';
import ShellCheckProvider from './dockerLinter';

export function activate(context: vscode.ExtensionContext) {
	console.log('Congratulations, your extension "docker-lint" is now active!');
	const linter = new ShellCheckProvider(context);
	context.subscriptions.push(linter);
}

// this method is called when your extension is deactivated
export function deactivate() {}
