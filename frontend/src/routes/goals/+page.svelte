<script lang="ts">
	import { enhance } from '$app/forms';
	import { formatMoney, formatDate, goalProgress, type Goal } from '$lib/api';

	let { data, form } = $props();

	let showCreate = $state(false);

	const statusColors: Record<string, string> = {
		active: 'bg-blue-100 text-blue-700',
		completed: 'bg-emerald-100 text-emerald-700',
		archived: 'bg-gray-100 text-gray-600'
	};
</script>

<svelte:head><title>Goals — FinGoat</title></svelte:head>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-2xl font-bold text-gray-900">Goals</h1>
		<button
			onclick={() => (showCreate = !showCreate)}
			class="rounded-lg bg-emerald-600 text-white px-4 py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
		>
			{showCreate ? 'Cancel' : '+ New Goal'}
		</button>
	</div>

	{#if showCreate}
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<h2 class="text-base font-semibold text-gray-900 mb-4">Create Goal</h2>
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
					<label for="title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
					<input
						name="title"
						required
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						placeholder="Emergency fund"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Target Amount</label>
					<input
						name="targetAmount"
						required
						type="number"
						step="0.01"
						min="0"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
						placeholder="5000.00"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Deadline (optional)</label>
					<input
						name="deadline"
						type="date"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
					/>
				</div>
				<div class="sm:col-span-3">
					<button
						type="submit"
						class="rounded-lg bg-emerald-600 text-white px-6 py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
					>
						Create Goal
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if data.goals.length === 0}
		<div class="bg-white rounded-xl border border-gray-200 p-10 text-center">
			<p class="text-gray-400">No goals yet. Create your first financial goal!</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
			{#each data.goals as goal (goal.id)}
				<div class="bg-white rounded-xl border border-gray-200 p-5 flex flex-col gap-3">
					<div class="flex justify-between items-start">
						<div>
							<h3 class="font-semibold text-gray-900">{goal.title}</h3>
							{#if goal.deadline}
								<p class="text-xs text-gray-400 mt-0.5">Due {formatDate(goal.deadline)}</p>
							{/if}
						</div>
						<span class="text-xs font-medium px-2 py-1 rounded-full {statusColors[goal.status] ?? statusColors.active}">
							{goal.status}
						</span>
					</div>

					<div>
						<div class="flex justify-between text-sm mb-1">
							<span class="text-gray-500">Progress</span>
							<span class="font-medium text-gray-900">
								{formatMoney(goal.currentAmount)} / {formatMoney(goal.targetAmount)}
							</span>
						</div>
						<div class="h-2 bg-gray-100 rounded-full overflow-hidden">
							<div
								class="h-full bg-emerald-500 rounded-full"
								style="width: {goalProgress(goal)}%"
							></div>
						</div>
						<p class="text-xs text-gray-400 mt-0.5">{goalProgress(goal).toFixed(0)}% complete</p>
					</div>

					<div class="flex gap-2 mt-auto">
						<a
							href="/goals/{goal.id}"
							class="flex-1 text-center rounded-lg border border-gray-200 text-sm py-1.5 text-gray-700 hover:bg-gray-50 transition-colors"
						>
							View / Edit
						</a>
						<form method="POST" action="?/delete" use:enhance>
							<input type="hidden" name="id" value={goal.id} />
							<button
								type="submit"
								onclick={(e) => { if (!confirm('Delete this goal?')) e.preventDefault(); }}
								class="rounded-lg border border-red-200 text-red-600 text-sm px-3 py-1.5 hover:bg-red-50 transition-colors"
							>
								Delete
							</button>
						</form>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
