<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	let { data, children } = $props();

	const navLinks = [
		{
			href: '/',
			label: 'Dashboard',
			icon: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`
		},
		{
			href: '/goals',
			label: 'Goals',
			icon: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/></svg>`
		},
		{
			href: '/transactions',
			label: 'Transactions',
			icon: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>`
		},
		{
			href: '/settings',
			label: 'Settings',
			icon: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`
		}
	];
</script>

{#if data.user}
	<div class="min-h-screen bg-gray-50 flex flex-col">
		<!-- Top nav -->
		<nav class="bg-white border-b border-gray-200 px-4 md:px-6 py-3 flex items-center justify-between sticky top-0 z-20">
			<span class="font-bold text-xl text-emerald-600">FinGoat</span>
			<!-- Desktop nav links -->
			<div class="hidden md:flex items-center gap-6">
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
			<!-- Mobile: compact user indicator -->
			<span class="md:hidden text-xs text-gray-400 truncate max-w-[180px]">{data.user.email}</span>
		</nav>

		<!-- Page content -->
		<main class="flex-1 max-w-5xl w-full mx-auto px-4 md:px-6 py-6 pb-24 md:pb-8">
			{@render children()}
		</main>

		<!-- Mobile bottom tab bar -->
		<nav class="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-20">
			<div class="flex items-stretch">
				{#each navLinks as link}
					<a
						href={link.href}
						class="flex-1 flex flex-col items-center justify-center gap-0.5 py-2 text-xs font-medium transition-colors {$page.url.pathname === link.href
							? 'text-emerald-600'
							: 'text-gray-400'}"
					>
						{@html link.icon}
						{link.label}
					</a>
				{/each}
				<form method="POST" action="/logout" class="flex-1">
					<button
						type="submit"
						class="w-full h-full flex flex-col items-center justify-center gap-0.5 py-2 text-xs font-medium text-gray-400 hover:text-red-500 transition-colors"
					>
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5">
							<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
							<polyline points="16 17 21 12 16 7"/>
							<line x1="21" y1="12" x2="9" y2="12"/>
						</svg>
						Logout
					</button>
				</form>
			</div>
		</nav>
	</div>
{:else}
	{@render children()}
{/if}
