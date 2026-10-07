import path from 'node:path';

const libRoot = path.join('src', 'lib');

const allowedTargets = {
	routes: ['features', 'components', 'api', 'i18n', 'styles', 'server'],
	features: ['components', 'api', 'domain', 'offline', 'i18n'],
	components: ['styles', 'i18n'],
	domain: [],
	offline: ['api', 'domain']
};

function layerOfFile(filename) {
	const relative = path.relative(process.cwd(), filename);
	if (relative.startsWith(path.join('src', 'routes'))) {
		return { name: 'routes', unit: 'routes' };
	}
	return layerOfLibPath(path.relative(libRoot, relative));
}

function layerOfLibPath(libPath) {
	if (libPath.startsWith('..')) {
		return undefined;
	}
	const [name, unit] = libPath.split(path.sep);
	const unitKey = name === 'features' ? `${name}/${unit}` : name;
	return { name, unit: unitKey };
}

function layerOfImport(specifier, filename) {
	const aliased = /^[#$]lib\/(.*)$/u.exec(specifier);
	if (aliased) {
		return layerOfLibPath(path.normalize(aliased[1]));
	}
	if (!specifier.startsWith('.')) {
		return undefined;
	}
	const absolute = path.resolve(path.dirname(filename), specifier);
	return layerOfLibPath(path.relative(path.resolve(libRoot), absolute));
}

function isAllowed(source, target) {
	const allowed = allowedTargets[source.name];
	if (!allowed || !target || source.unit === target.unit) {
		return true;
	}
	return allowed.includes(target.name);
}

export const layerImports = {
	meta: {
		type: 'problem',
		schema: [],
		messages: { forbidden: 'Layer "{{source}}" must not import from "{{target}}".' }
	},
	create(context) {
		const source = layerOfFile(context.filename);
		return {
			ImportDeclaration(node) {
				const target = layerOfImport(node.source.value, context.filename);
				if (source && !isAllowed(source, target)) {
					context.report({
						node,
						messageId: 'forbidden',
						data: { source: source.unit, target: target.unit }
					});
				}
			}
		};
	}
};
