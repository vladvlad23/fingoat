import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { api } from '$lib/api';

function getCycleStart(paymentDay: number): Date {
	const today = new Date();
	const day = today.getDate();
	if (day >= paymentDay) {
		return new Date(today.getFullYear(), today.getMonth(), paymentDay);
	} else {
		return new Date(today.getFullYear(), today.getMonth() - 1, paymentDay);
	}
}

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.token || !locals.user) redirect(302, '/login');

	const [goals, transactions] = await Promise.all([
		api.goals.list(fetch, locals.token),
		api.transactions.list(fetch, locals.token)
	]);

	let cycleSpending: string | null = null;
	if (locals.user.paymentDay) {
		const cycleStart = getCycleStart(locals.user.paymentDay);
		const cycleStartStr = cycleStart.toISOString().split('T')[0];
		const spent = transactions
			.filter((tx) => tx.type === 'expense' && tx.date >= cycleStartStr)
			.reduce((sum, tx) => sum + parseFloat(tx.amount), 0);
		cycleSpending = spent.toFixed(2);
	}

	return {
		user: locals.user,
		goals,
		recentTransactions: transactions.slice(0, 5),
		cycleSpending
	};
};
