const message = 'Comments are not allowed. Use a better name, an extracted function, or an ADR.';

export const noComments = {
	meta: {
		type: 'problem',
		schema: [],
		messages: { comment: message }
	},
	create(context) {
		const report = (node) => context.report({ loc: node.loc, messageId: 'comment' });
		return {
			Program() {
				context.sourceCode
					.getAllComments()
					.filter((comment) => comment.type !== 'Shebang')
					.forEach(report);
			},
			SvelteHTMLComment: report
		};
	}
};
