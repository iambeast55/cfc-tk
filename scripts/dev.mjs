import { spawn } from 'node:child_process';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const root = new URL('../', import.meta.url);
const env = {
	...process.env,
	GOCACHE: process.env.GOCACHE ?? fileURLToPath(new URL('.gocache/', root))
};

const commands = [
	{
		name: 'server',
		command: 'go run -buildvcs=false .',
		cwd: new URL('../server/', import.meta.url)
	},
	{
		name: 'ui',
		command: 'npm run dev:ui -- --host 127.0.0.1',
		cwd: new URL('../', import.meta.url)
	}
];

const children = [];
let shuttingDown = false;

function prefixStream(stream, name, output) {
	let buffer = '';

	stream.on('data', (chunk) => {
		buffer += chunk.toString();
		const lines = buffer.split(/\r?\n/);
		buffer = lines.pop() ?? '';

		for (const line of lines) {
			if (line.length > 0) {
				output.write(`[${name}] ${line}\n`);
			}
		}
	});

	stream.on('end', () => {
		if (buffer.length > 0) {
			output.write(`[${name}] ${buffer}\n`);
		}
	});
}

function stopChild(child) {
	if (!child.pid || child.killed) return;

	if (process.platform === 'win32') {
		spawn('taskkill', ['/pid', String(child.pid), '/t', '/f'], {
			stdio: 'ignore'
		});
		return;
	}

	child.kill('SIGTERM');
}

function shutdown(exitCode = 0) {
	if (shuttingDown) return;
	shuttingDown = true;

	for (const child of children) {
		stopChild(child);
	}

	setTimeout(() => process.exit(exitCode), 250);
}

for (const { name, command, cwd } of commands) {
	const child = spawn(command, {
		cwd,
		shell: true,
		stdio: ['inherit', 'pipe', 'pipe'],
		env
	});

	children.push(child);
	prefixStream(child.stdout, name, process.stdout);
	prefixStream(child.stderr, name, process.stderr);

	child.on('exit', (code, signal) => {
		if (shuttingDown) return;

		if (code === 0 || signal === 'SIGTERM' || signal === 'SIGINT') {
			shutdown(0);
			return;
		}

		console.error(`[${name}] exited with code ${code ?? signal}`);
		shutdown(code ?? 1);
	});
}

process.on('SIGINT', () => shutdown(0));
process.on('SIGTERM', () => shutdown(0));
