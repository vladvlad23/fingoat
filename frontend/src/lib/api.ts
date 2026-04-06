import { PUBLIC_API_URL } from '$env/static/public';

export interface User {
	id: number;
	email: string;
	monthlyIncome: string;
	paymentDay?: number;
	createdAt: string;
}

export interface Goal {
	id: number;
	userId: number;
	title: string;
	targetAmount: string;
	currentAmount: string;
	deadline?: string;
	status: string;
	createdAt: string;
}

export interface Transaction {
	id: number;
	userId: number;
	goalId?: number;
	title: string;
	amount: string;
	type: string;
	category?: string;
	date: string;
	createdAt: string;
}

export interface ApiError {
	error: string;
}

const BASE = PUBLIC_API_URL;

async function request<T>(
	fetch: typeof globalThis.fetch,
	method: string,
	path: string,
	token: string | null,
	body?: unknown
): Promise<T> {
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(`${BASE}${path}`, {
		method,
		headers,
		body: body !== undefined ? JSON.stringify(body) : undefined
	});

	if (res.status === 204) return undefined as T;

	const data = await res.json();
	if (!res.ok) throw new Error((data as ApiError).error ?? 'Request failed');
	return data as T;
}

export const api = {
	auth: {
		register: (fetch: typeof globalThis.fetch, email: string, password: string) =>
			request<{ token: string }>(fetch, 'POST', '/api/auth/register', null, { email, password }),
		login: (fetch: typeof globalThis.fetch, email: string, password: string) =>
			request<{ token: string }>(fetch, 'POST', '/api/auth/login', null, { email, password }),
		/** Exchange the httpOnly refresh_token cookie for a new access token. */
		refresh: (fetch: typeof globalThis.fetch) =>
			request<{ token: string }>(fetch, 'POST', '/api/auth/refresh', null),
		/** Revoke the refresh token on the backend. */
		logout: (fetch: typeof globalThis.fetch) =>
			request<void>(fetch, 'POST', '/api/auth/logout', null)
	},
	user: {
		me: (fetch: typeof globalThis.fetch, token: string) =>
			request<User>(fetch, 'GET', '/api/me', token),
		updateIncome: (fetch: typeof globalThis.fetch, token: string, monthlyIncome: string) =>
			request<User>(fetch, 'PATCH', '/api/me', token, { monthlyIncome }),
		updateSettings: (fetch: typeof globalThis.fetch, token: string, monthlyIncome: string, paymentDay?: number) =>
			request<User>(fetch, 'PATCH', '/api/me', token, { monthlyIncome, paymentDay })
	},
	goals: {
		list: (fetch: typeof globalThis.fetch, token: string) =>
			request<Goal[]>(fetch, 'GET', '/api/goals', token),
		get: (fetch: typeof globalThis.fetch, token: string, id: number) =>
			request<Goal>(fetch, 'GET', `/api/goals/${id}`, token),
		create: (
			fetch: typeof globalThis.fetch,
			token: string,
			data: { title: string; targetAmount: string; deadline?: string }
		) => request<Goal>(fetch, 'POST', '/api/goals', token, data),
		update: (
			fetch: typeof globalThis.fetch,
			token: string,
			id: number,
			data: Partial<{ title: string; targetAmount: string; deadline: string; status: string }>
		) => request<Goal>(fetch, 'PUT', `/api/goals/${id}`, token, data),
		delete: (fetch: typeof globalThis.fetch, token: string, id: number) =>
			request<void>(fetch, 'DELETE', `/api/goals/${id}`, token)
	},
	transactions: {
		list: (
			fetch: typeof globalThis.fetch,
			token: string,
			params?: { goalId?: number; type?: string; from?: string; to?: string }
		) => {
			const qs = params
				? '?' +
					new URLSearchParams(
						Object.entries(params)
							.filter(([, v]) => v !== undefined)
							.map(([k, v]) => [k, String(v)])
					).toString()
				: '';
			return request<Transaction[]>(fetch, 'GET', `/api/transactions${qs}`, token);
		},
		get: (fetch: typeof globalThis.fetch, token: string, id: number) =>
			request<Transaction>(fetch, 'GET', `/api/transactions/${id}`, token),
		create: (
			fetch: typeof globalThis.fetch,
			token: string,
			data: {
				title: string;
				amount: string;
				type: string;
				date: string;
				goalId?: number;
				category?: string;
			}
		) => request<Transaction>(fetch, 'POST', '/api/transactions', token, data),
		update: (
			fetch: typeof globalThis.fetch,
			token: string,
			id: number,
			data: Partial<{
				title: string;
				amount: string;
				type: string;
				date: string;
				goalId: number;
				category: string;
			}>
		) => request<Transaction>(fetch, 'PUT', `/api/transactions/${id}`, token, data),
		delete: (fetch: typeof globalThis.fetch, token: string, id: number) =>
			request<void>(fetch, 'DELETE', `/api/transactions/${id}`, token)
	}
};

export function formatMoney(s: string): string {
	const n = parseFloat(s);
	if (isNaN(n)) return s;
	return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);
}

export function formatDate(s: string): string {
	if (!s) return '';
	return new Date(s + 'T00:00:00').toLocaleDateString('en-US', {
		year: 'numeric',
		month: 'short',
		day: 'numeric'
	});
}

export function goalProgress(goal: Goal): number {
	const current = parseFloat(goal.currentAmount);
	const target = parseFloat(goal.targetAmount);
	if (target <= 0) return 0;
	return Math.min(100, (current / target) * 100);
}
