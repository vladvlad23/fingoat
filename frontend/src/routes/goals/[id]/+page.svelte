<script lang="ts">
	import { enhance } from '$app/forms';
	import { formatMoney, formatDate, goalProgress } from '$lib/api';

	let { data, form } = $props();

	let editing = $state(false);

	const statusOptions = ['active', 'completed', 'archived'];
	const statusColors: Record<string, string> = {
		active: 'bg-blue-100 text-blue-700',
		completed: 'bg-emerald-100 text-emerald-700',
		archived: 'bg-gray-100 text-gray-600'
	};
</script>

<svelte:head><title>{data.goal.title} — FinGoat</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center gap-2">
		<a href="/goals" class="text-sm text-gray-400 hover:text-gray-600">← Goals</a>
	</div>

	<div class="bg-white rounded-xl border border-gray-200 p-6">
		<div class="flex justify-between items-start mb-4">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">{data.goal.title}</h1>
				{#if data.goal.deadline}
					<p class="text-sm text-gray-500 mt-0.5">Due {formatDate(data.goal.deadline)}</p>
				{/if}
			</div>
			<div class="flex items-center gap-2">
				<span class="text-xs font-medium px-2 py-1 rounded-full {statusColors[data.goal.status] ?? statusColors.active}">
					{data.goal.status}
				</span>
				<button
					onclick={() => (editing = !editing)}
					class="text-sm text-emerald-600 hover:underline"
				>
					{editing ? 'Cancel' : 'Edit'}
				</button>
			</div>
		</div>

		<!-- Progress -->
		<div class="mb-6">
			<div class="flex justify-between text-sm mb-2">
				<span class="text-gray-500">Progress</span>
				<span class="font-semibold text-gray-900">
					{formatMoney(data.goal.currentAmount, data.goal.currency)} / {formatMoney(data.goal.targetAmount, data.goal.currency)}
				</span>
			</div>
			<div class="h-3 bg-gray-100 rounded-full overflow-hidden">
				<div
					class="h-full bg-emerald-500 rounded-full transition-all"
					style="width: {goalProgress(data.goal)}%"
				></div>
			</div>
			<p class="text-xs text-gray-400 mt-1">{goalProgress(data.goal).toFixed(1)}% complete</p>
		</div>

		{#if editing}
			<div class="border-t border-gray-100 pt-4">
				<h2 class="text-sm font-semibold text-gray-700 mb-3">Edit Goal</h2>
				{#if form?.updateError}
					<div class="mb-3 rounded-lg bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">
						{form.updateError}
					</div>
				{/if}
				<form
					method="POST"
					action="?/update"
					use:enhance={() => {
						return ({ result }) => {
							if (result.type === 'success' || result.type === 'redirect') editing = false;
						};
					}}
					class="grid grid-cols-1 sm:grid-cols-2 gap-4"
				>
					<div>
						<label for="edit-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
						<input
							id="edit-title"
							name="title"
							value={data.goal.title}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						/>
					</div>
					<div>
						<label for="edit-targetAmount" class="block text-sm font-medium text-gray-700 mb-1">Target Amount</label>
						<input
							id="edit-targetAmount"
							name="targetAmount"
							value={data.goal.targetAmount}
							type="number"
							step="0.01"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						/>
					</div>
					<div>
						<label for="edit-deadline" class="block text-sm font-medium text-gray-700 mb-1">Deadline</label>
						<input
							id="edit-deadline"
							name="deadline"
							type="date"
							value={data.goal.deadline ?? ''}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						/>
					</div>
					<div>
						<label for="edit-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
						<select
							id="edit-status"
							name="status"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						>
							{#each statusOptions as s}
								<option value={s} selected={s === data.goal.status}>{s}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="edit-currency" class="block text-sm font-medium text-gray-700 mb-1">Currency</label>
						<select
							id="edit-currency"
							name="currency"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						>
							{#each ['USD', 'EUR', 'RON'] as c}
								<option value={c} selected={c === data.goal.currency}>{c}</option>
							{/each}
						</select>
					</div>
					<div class="sm:col-span-2">
						<button
							type="submit"
							class="rounded-lg bg-emerald-600 text-white px-6 py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
						>
							Save Changes
						</button>
					</div>
				</form>
			</div>
		{/if}
	</div>

	<!-- Linked transactions -->
	<div class="bg-white rounded-xl border border-gray-200 p-6">
		<div class="flex justify-between items-center mb-4">
			<h2 class="text-base font-semibold text-gray-900">Linked Transactions</h2>
			<a
				href="/transactions?goalId={data.goal.id}"
				class="text-sm text-emerald-600 hover:underline"
			>
				View in Transactions
			</a>
		</div>
		{#if data.transactions.length === 0}
			<p class="text-sm text-gray-400">No transactions linked to this goal yet.</p>
		{:else}
			<div class="divide-y divide-gray-100">
				{#each data.transactions as tx (tx.id)}
					<div class="py-3 flex justify-between items-center">
						<div>
							<p class="text-sm font-medium text-gray-900">{tx.title}</p>
							<p class="text-xs text-gray-400">{formatDate(tx.date)}</p>
						</div>
						<span
							class="text-sm font-semibold {tx.type === 'income' ? 'text-emerald-600' : 'text-red-500'}"
						>
							{tx.type === 'income' ? '+' : '-'}{formatMoney(tx.amount, tx.currency)}
						</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
