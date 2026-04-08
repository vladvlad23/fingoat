<script lang="ts">
	import { formatMoney, formatDate, goalProgress, type Goal } from '$lib/api';

	let { data } = $props();

	const activeGoals = $derived(data.goals.filter((g: Goal) => g.status !== 'archived'));
	const completedGoals = $derived(data.goals.filter((g: Goal) => g.status === 'completed'));
</script>

<svelte:head><title>Dashboard — FinGoat</title></svelte:head>

<div class="space-y-8">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
		<p class="text-gray-500 mt-1">Welcome back, {data.user.email}</p>
	</div>

	<!-- Summary cards -->
	<div class="grid grid-cols-1 sm:grid-cols-2 {data.cycleSpending !== null ? 'md:grid-cols-4' : 'md:grid-cols-3'} gap-4">
		<div class="bg-white rounded-xl border border-gray-200 p-5">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wide">Monthly Income</p>
			<p class="text-2xl font-bold text-gray-900 mt-1">{formatMoney(data.user.monthlyIncome, data.user.currency)}</p>
		</div>
		<div class="bg-white rounded-xl border border-gray-200 p-5">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wide">Active Goals</p>
			<p class="text-2xl font-bold text-gray-900 mt-1">{activeGoals.length}</p>
		</div>
		<div class="bg-white rounded-xl border border-gray-200 p-5">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wide">Completed Goals</p>
			<p class="text-2xl font-bold text-emerald-600 mt-1">{completedGoals.length}</p>
		</div>
		{#if data.cycleSpending !== null}
		<div class="bg-white rounded-xl border border-gray-200 p-5">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wide">This Cycle</p>
			<p class="text-2xl font-bold text-red-500 mt-1">{formatMoney(data.cycleSpending, data.user.currency)}</p>
			<p class="text-xs text-gray-400 mt-1">expenses since day {data.user.paymentDay}</p>
		</div>
		{/if}
	</div>

	<!-- Goals progress -->
	{#if activeGoals.length > 0}
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<h2 class="text-base font-semibold text-gray-900 mb-4">Goals Progress</h2>
			<div class="space-y-4">
				{#each activeGoals.slice(0, 4) as goal (goal.id)}
					<div>
						<div class="flex justify-between items-baseline gap-2 text-sm mb-1">
							<span class="font-medium text-gray-800 truncate min-w-0">{goal.title}</span>
							<span class="text-gray-500 whitespace-nowrap shrink-0 text-xs"
								>{formatMoney(goal.currentAmount, goal.currency)} / {formatMoney(goal.targetAmount, goal.currency)}</span
							>
						</div>
						<div class="h-2 bg-gray-100 rounded-full overflow-hidden">
							<div
								class="h-full bg-emerald-500 rounded-full transition-all"
								style="width: {goalProgress(goal)}%"
							></div>
						</div>
						{#if goal.deadline}
							<p class="text-xs text-gray-400 mt-0.5">Deadline: {formatDate(goal.deadline)}</p>
						{/if}
					</div>
				{/each}
			</div>
			{#if activeGoals.length > 4}
				<a href="/goals" class="text-sm text-emerald-600 hover:underline mt-3 inline-block">
					View all {activeGoals.length} goals →
				</a>
			{/if}
		</div>
	{/if}

	<!-- Recent transactions -->
	<div class="bg-white rounded-xl border border-gray-200 p-6">
		<div class="flex justify-between items-center mb-4">
			<h2 class="text-base font-semibold text-gray-900">Recent Transactions</h2>
			<a href="/transactions" class="text-sm text-emerald-600 hover:underline">View all</a>
		</div>
		{#if data.recentTransactions.length === 0}
			<p class="text-sm text-gray-400">No transactions yet.</p>
		{:else}
			<div class="divide-y divide-gray-100">
				{#each data.recentTransactions as tx (tx.id)}
					<div class="py-3 flex justify-between items-center">
						<div>
							<p class="text-sm font-medium text-gray-900">{tx.title}</p>
							<p class="text-xs text-gray-400">{formatDate(tx.date)}{tx.category ? ` · ${tx.category}` : ''}</p>
						</div>
						<span
							class="text-sm font-semibold {tx.type === 'income'
								? 'text-emerald-600'
								: 'text-red-500'}"
						>
							{tx.type === 'income' ? '+' : '-'}{formatMoney(tx.amount, tx.currency)}
						</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
