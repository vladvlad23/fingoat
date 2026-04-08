import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { api } from '$lib/api';

export const load: PageServerLoad = async ({ locals, fetch, url }) => {
	if (!locals.token || !locals.user) redirect(302, '/login');

	const goalId = url.searchParams.get('goalId');
	const type = url.searchParams.get('type');
	const from = url.searchParams.get('from');
	const to = url.searchParams.get('to');

	const [transactions, goals] = await Promise.all([
		api.transactions.list(fetch, locals.token, {
			...(goalId && { goalId: Number(goalId) }),
			...(type && { type }),
			...(from && { from }),
			...(to && { to })
		}),
		api.goals.list(fetch, locals.token)
	]);

	return { transactions, goals, filters: { goalId, type, from, to } };
};

export const actions: Actions = {
	create: async ({ request, locals, fetch }) => {
		if (!locals.token) return fail(401, { createError: 'Unauthorized' });
		const form = await request.formData();
		const title = (form.get('title') as string)?.trim();
		const amount = (form.get('amount') as string)?.trim();
		const type = form.get('type') as string;
		const date = form.get('date') as string;
		const category = (form.get('category') as string)?.trim() || undefined;
		const goalIdStr = form.get('goalId') as string;
		const goalId = goalIdStr ? Number(goalIdStr) : undefined;
		const currency = (form.get('currency') as string)?.trim() || undefined;

		if (!title || !amount || !type || !date)
			return fail(400, { createError: 'Title, amount, type, and date are required' });

		try {
			await api.transactions.create(fetch, locals.token, {
				title,
				amount,
				type,
				date,
				...(category && { category }),
				...(goalId && { goalId }),
				...(currency && { currency })
			});
		} catch (e) {
			return fail(400, { createError: (e as Error).message });
		}
	},
	delete: async ({ request, locals, fetch }) => {
		if (!locals.token) return fail(401, { error: 'Unauthorized' });
		const form = await request.formData();
		const id = Number(form.get('id'));
		try {
			await api.transactions.delete(fetch, locals.token, id);
		} catch (e) {
			return fail(400, { error: (e as Error).message });
		}
	},
	update: async ({ request, locals, fetch }) => {
		if (!locals.token) return fail(401, { updateError: 'Unauthorized' });
		const form = await request.formData();
		const id = Number(form.get('id'));
		const title = (form.get('title') as string)?.trim();
		const amount = (form.get('amount') as string)?.trim();
		const type = form.get('type') as string;
		const date = form.get('date') as string;
		const category = (form.get('category') as string)?.trim() || undefined;
		const goalIdStr = form.get('goalId') as string;
		const goalId = goalIdStr ? Number(goalIdStr) : undefined;
		const txCurrency = (form.get('currency') as string)?.trim() || undefined;

		try {
			await api.transactions.update(fetch, locals.token, id, {
				...(title && { title }),
				...(amount && { amount }),
				...(type && { type }),
				...(date && { date }),
				...(category !== undefined && { category }),
				...(goalId !== undefined && { goalId }),
				...(txCurrency && { currency: txCurrency })
			});
		} catch (e) {
			return fail(400, { updateError: (e as Error).message });
		}
	}
};
