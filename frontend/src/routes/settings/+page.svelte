<script lang="ts">
	import { enhance } from '$app/forms';
	import { formatMoney } from '$lib/api';

	let { data, form } = $props();
</script>

<svelte:head><title>Settings — FinGoat</title></svelte:head>

<div class="max-w-lg space-y-6">
	<h1 class="text-2xl font-bold text-gray-900">Settings</h1>

	<div class="bg-white rounded-xl border border-gray-200 p-6">
		<h2 class="text-base font-semibold text-gray-900 mb-1">Account</h2>
		<p class="text-sm text-gray-500 mb-4">{data.user.email}</p>

		<h3 class="text-sm font-semibold text-gray-700 mb-3">Monthly Income</h3>
		<p class="text-sm text-gray-500 mb-4">
			Current: <span class="font-medium text-gray-900">{formatMoney(data.user.monthlyIncome)}</span>
		</p>
		<p class="text-sm text-gray-500 mb-4">
			Payment day: <span class="font-medium text-gray-900">
				{data.user.paymentDay ? `Day ${data.user.paymentDay}` : 'Not set'}
			</span>
		</p>

		{#if form?.error}
			<div class="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">
				{form.error}
			</div>
		{/if}
		{#if form?.success}
			<div class="mb-4 rounded-lg bg-emerald-50 border border-emerald-200 px-4 py-2 text-sm text-emerald-700">
				Settings updated successfully.
			</div>
		{/if}

		<form method="POST" use:enhance class="flex gap-3 items-end">
			<div class="flex-1">
				<label for="monthlyIncome" class="block text-sm font-medium text-gray-700 mb-1">
					New Monthly Income
				</label>
				<input
					id="monthlyIncome"
					name="monthlyIncome"
					type="number"
					step="0.01"
					min="0"
					required
					value={form?.user?.monthlyIncome ?? data.user.monthlyIncome}
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					placeholder="5000.00"
				/>
			</div>
			<div class="flex-1">
				<label for="paymentDay" class="block text-sm font-medium text-gray-700 mb-1">
					Payment Day (1–28, optional)
				</label>
				<input
					id="paymentDay"
					name="paymentDay"
					type="number"
					min="1"
					max="28"
					value={form?.user?.paymentDay ?? data.user.paymentDay ?? ''}
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					placeholder="25"
				/>
			</div>
			<button
				type="submit"
				class="rounded-lg bg-emerald-600 text-white px-5 py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
			>
				Update
			</button>
		</form>
	</div>

	<div class="bg-white rounded-xl border border-gray-200 p-6">
		<h2 class="text-base font-semibold text-gray-900 mb-3">Danger Zone</h2>
		<form method="POST" action="/logout">
			<button
				type="submit"
				class="rounded-lg border border-red-200 text-red-600 px-4 py-2 text-sm font-medium hover:bg-red-50 transition-colors"
			>
				Sign Out
			</button>
		</form>
	</div>
</div>
