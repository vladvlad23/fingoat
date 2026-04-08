// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Locals {
			user: { id: number; email: string; monthlyIncome: string; paymentDay?: number; currency: string } | null;
			token: string | null;
		}
	}
}

export {};
