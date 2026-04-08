<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { formatMoney, formatDate, type Goal } from '$lib/api';
	import ConfirmDialog from '$lib/ConfirmDialog.svelte';

	let { data, form } = $props();

	let showCreate = $state(false);
	let editingId = $state<number | null>(null);
	let confirmOpen = $state(false);
	let pendingDeleteForm = $state<HTMLFormElement | null>(null);
	let newTxDate = $state(new Date().toISOString().split('T')[0]);

	// filters
	let filterGoalId = $state($page.url.searchParams.get('goalId') ?? '');
	let filterType = $state($page.url.searchParams.get('type') ?? '');
	let filterFrom = $state($page.url.searchParams.get('from') ?? '');
	let filterTo = $state($page.url.searchParams.get('to') ?? '');

	function applyFilters() {
		const params = new URLSearchParams();
		if (filterGoalId) params.set('goalId', filterGoalId);
		if (filterType) params.set('type', filterType);
		if (filterFrom) params.set('from', filterFrom);
		if (filterTo) params.set('to', filterTo);
		goto(`/transactions?${params.toString()}`);
	}

	function clearFilters() {
		filterGoalId = '';
		filterType = '';
		filterFrom = '';
		filterTo = '';
		goto('/transactions');
	}

	const goalMap = $derived(
		Object.fromEntries((data.goals as Goal[]).map((g) => [g.id, g.title]))
	);
</script>

<svelte:head><title>Transactions — FinGoat</title></svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-2xl font-bold text-gray-900">Transactions</h1>
		<button
			onclick={() => (showCreate = !showCreate)}
			class="rounded-lg bg-emerald-600 text-white px-4 py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
		>
			{showCreate ? 'Cancel' : '+ New Transaction'}
		</button>
	</div>

	{#if showCreate}
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<h2 class="text-base font-semibold text-gray-900 mb-4">New Transaction</h2>
			{#if form?.createError}
				<div class="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">
					{form.createError}
				</div>
			{/if}
			<form
				method="POST"
				action="?/create"
				use:enhance={() => {
					return ({ result, update }) => {
						if (result.type === 'success') showCreate = false;
						update();
					};
				}}
				class="grid grid-cols-1 sm:grid-cols-3 gap-4"
			>
				<div>
					<label for="new-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
					<input
						id="new-title"
						name="title"
						required
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					/>
				</div>
				<div>
					<label for="new-amount" class="block text-sm font-medium text-gray-700 mb-1">Amount</label>
					<input
						id="new-amount"
						name="amount"
						type="number"
						step="0.01"
						min="0"
						required
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					/>
				</div>
				<div>
					<label for="new-type" class="block text-sm font-medium text-gray-700 mb-1">Type</label>
					<select
						id="new-type"
						name="type"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					>
						<option value="expense">Expense</option>
						<option value="income">Income</option>
					</select>
				</div>
				<div>
					<label for="new-date" class="block text-sm font-medium text-gray-700 mb-1">Date</label>
					<input
						id="new-date"
						name="date"
						type="date"
						required
						bind:value={newTxDate}
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					/>
				</div>
				<div>
					<label for="new-category" class="block text-sm font-medium text-gray-700 mb-1">Category (optional)</label>
					<input
						id="new-category"
						name="category"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						placeholder="Food, transport..."
					/>
				</div>
				<div>
					<label for="new-goalId" class="block text-sm font-medium text-gray-700 mb-1">Goal (optional)</label>
					<select
						id="new-goalId"
						name="goalId"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					>
						<option value="">None</option>
						{#each data.goals as goal (goal.id)}
							<option value={goal.id}>{goal.title}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="new-currency" class="block text-sm font-medium text-gray-700 mb-1">Currency</label>
					<select
						id="new-currency"
						name="currency"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					>
						{#each ['RON', 'EUR', 'USD'] as c}
							<option value={c}>{c}</option>
						{/each}
					</select>
				</div>
				<div class="sm:col-span-3">
					<button
						type="submit"
						class="rounded-lg bg-emerald-600 text-white px-6 py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
					>
						Add Transaction
					</button>
				</div>
			</form>
		</div>
	{/if}

	<!-- Filters -->
	<div class="bg-white rounded-xl border border-gray-200 p-4">
		<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 items-end">
			<div>
				<label for="filter-goalId" class="block text-xs font-medium text-gray-500 mb-1">Goal</label>
				<select
					id="filter-goalId"
					bind:value={filterGoalId}
					class="w-full rounded-lg border border-gray-200 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
				>
					<option value="">All goals</option>
					{#each data.goals as goal (goal.id)}
						<option value={String(goal.id)}>{goal.title}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="filter-type" class="block text-xs font-medium text-gray-500 mb-1">Type</label>
				<select
					id="filter-type"
					bind:value={filterType}
					class="w-full rounded-lg border border-gray-200 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
				>
					<option value="">All types</option>
					<option value="income">Income</option>
					<option value="expense">Expense</option>
				</select>
			</div>
			<div>
				<label for="filter-from" class="block text-xs font-medium text-gray-500 mb-1">From</label>
				<input
					id="filter-from"
					type="date"
					bind:value={filterFrom}
					class="w-full rounded-lg border border-gray-200 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
				/>
			</div>
			<div>
				<label for="filter-to" class="block text-xs font-medium text-gray-500 mb-1">To</label>
				<input
					id="filter-to"
					type="date"
					bind:value={filterTo}
					class="w-full rounded-lg border border-gray-200 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
				/>
			</div>
		</div>
		<div class="flex gap-2 mt-3">
			<button
				onclick={applyFilters}
				class="rounded-lg bg-gray-800 text-white px-4 py-1.5 text-sm font-medium hover:bg-gray-700 transition-colors"
			>
				Apply
			</button>
			<button
				onclick={clearFilters}
				class="rounded-lg border border-gray-200 text-gray-600 px-4 py-1.5 text-sm font-medium hover:bg-gray-50 transition-colors"
			>
				Clear
			</button>
		</div>
	</div>

	<!-- Transaction list -->
	<div class="bg-white rounded-xl border border-gray-200">
		{#if data.transactions.length === 0}
			<div class="p-10 text-center">
				<p class="text-gray-400">No transactions found.</p>
			</div>
		{:else}
			<div class="divide-y divide-gray-100">
				{#each data.transactions as tx (tx.id)}
					{#if editingId === tx.id}
						<div class="p-4">
							{#if form?.updateError}
								<div class="mb-3 rounded-lg bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">
									{form.updateError}
								</div>
							{/if}
							<form
								method="POST"
								action="?/update"
								use:enhance={() => {
									return ({ result, update }) => {
										if (result.type === 'success') editingId = null;
										update();
									};
								}}
								class="grid grid-cols-2 sm:grid-cols-3 gap-3"
							>
								<input type="hidden" name="id" value={tx.id} />
								<div>
									<label for="edit-title-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Title</label>
									<input
										id="edit-title-{tx.id}"
										name="title"
										value={tx.title}
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									/>
								</div>
								<div>
									<label for="edit-amount-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Amount</label>
									<input
										id="edit-amount-{tx.id}"
										name="amount"
										type="number"
										step="0.01"
										value={tx.amount}
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									/>
								</div>
								<div>
									<label for="edit-type-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Type</label>
									<select
										id="edit-type-{tx.id}"
										name="type"
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									>
										<option value="expense" selected={tx.type === 'expense'}>Expense</option>
										<option value="income" selected={tx.type === 'income'}>Income</option>
									</select>
								</div>
								<div>
									<label for="edit-date-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Date</label>
									<input
										id="edit-date-{tx.id}"
										name="date"
										type="date"
										value={tx.date}
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									/>
								</div>
								<div>
									<label for="edit-category-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Category</label>
									<input
										id="edit-category-{tx.id}"
										name="category"
										value={tx.category ?? ''}
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									/>
								</div>
								<div>
									<label for="edit-goalId-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Goal</label>
									<select
										id="edit-goalId-{tx.id}"
										name="goalId"
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									>
										<option value="">None</option>
										{#each data.goals as goal (goal.id)}
											<option value={goal.id} selected={goal.id === tx.goalId}>{goal.title}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="edit-currency-{tx.id}" class="block text-xs font-medium text-gray-500 mb-1">Currency</label>
									<select
										id="edit-currency-{tx.id}"
										name="currency"
										class="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
									>
										{#each ['USD', 'EUR', 'RON'] as c}
											<option value={c} selected={c === tx.currency}>{c}</option>
										{/each}
									</select>
								</div>
								<div class="col-span-2 sm:col-span-3 flex gap-2">
									<button
										type="submit"
										class="rounded-lg bg-emerald-600 text-white px-4 py-1.5 text-sm font-semibold hover:bg-emerald-700 transition-colors"
									>
										Save
									</button>
									<button
										type="button"
										onclick={() => (editingId = null)}
										class="rounded-lg border border-gray-200 text-gray-600 px-4 py-1.5 text-sm hover:bg-gray-50 transition-colors"
									>
										Cancel
									</button>
								</div>
							</form>
						</div>
					{:else}
						<div class="px-5 py-3 flex items-center justify-between gap-4">
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium text-gray-900 truncate">{tx.title}</p>
								<p class="text-xs text-gray-400">
									{formatDate(tx.date)}
									{#if tx.category} · {tx.category}{/if}
									{#if tx.goalId} · <span class="text-emerald-600">{goalMap[tx.goalId]}</span>{/if}
								</p>
							</div>
							<span
								class="text-sm font-semibold whitespace-nowrap {tx.type === 'income'
									? 'text-emerald-600'
									: 'text-red-500'}"
							>
								{tx.type === 'income' ? '+' : '-'}{formatMoney(tx.amount, tx.currency)}
							</span>
							<div class="flex gap-2 shrink-0">
								<button
									onclick={() => (editingId = tx.id)}
									class="text-xs text-gray-400 hover:text-gray-700 transition-colors"
								>
									Edit
								</button>
								<form method="POST" action="?/delete" use:enhance>
									<input type="hidden" name="id" value={tx.id} />
									<button
										type="submit"
										onclick={(e) => {
											e.preventDefault();
											pendingDeleteForm = e.currentTarget.closest('form');
											confirmOpen = true;
										}}
										class="text-xs text-red-400 hover:text-red-600 transition-colors"
									>
										Delete
									</button>
								</form>
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{/if}
	</div>
</div>

<ConfirmDialog
	open={confirmOpen}
	title="Delete transaction?"
	message="This will permanently delete the transaction and cannot be undone."
	onconfirm={() => {
		pendingDeleteForm?.requestSubmit();
		confirmOpen = false;
	}}
	oncancel={() => {
		confirmOpen = false;
		pendingDeleteForm = null;
	}}
/>
