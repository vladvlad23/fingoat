<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	let { data, children } = $props();

	const navLinks = [
		{ href: '/', label: 'Dashboard' },
		{ href: '/goals', label: 'Goals' },
		{ href: '/transactions', label: 'Transactions' },
		{ href: '/settings', label: 'Settings' }
	];
</script>

{#if data.user}
	<div class="min-h-screen bg-gray-50 flex flex-col">
		<nav class="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
			<span class="font-bold text-xl text-emerald-600">FinGoat</span>
			<div class="flex items-center gap-6">
				{#each navLinks as link}
					<a
						href={link.href}
						class="text-sm font-medium transition-colors {$page.url.pathname === link.href
							? 'text-emerald-600'
							: 'text-gray-500 hover:text-gray-900'}"
					>
						{link.label}
					</a>
				{/each}
				<span class="text-xs text-gray-400">{data.user.email}</span>
				<form method="POST" action="/logout">
					<button class="text-sm text-gray-500 hover:text-red-600 transition-colors">Logout</button>
				</form>
			</div>
		</nav>
		<main class="flex-1 max-w-5xl w-full mx-auto px-6 py-8">
			{@render children()}
		</main>
	</div>
{:else}
	{@render children()}
{/if}
