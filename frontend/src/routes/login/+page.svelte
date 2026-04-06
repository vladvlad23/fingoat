<script lang="ts">
	import { enhance } from '$app/forms';

	let { form } = $props();
</script>

<svelte:head><title>Login — FinGoat</title></svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50">
	<div class="w-full max-w-sm bg-white rounded-2xl border border-gray-200 p-8 shadow-sm">
		<div class="mb-6 text-center">
			<span class="text-3xl font-bold text-emerald-600">FinGoat</span>
			<p class="text-gray-500 mt-1 text-sm">Sign in to your account</p>
		</div>

		{#if form?.error}
			<div class="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
				{form.error}
			</div>
		{/if}

		<form method="POST" use:enhance class="space-y-4">
			<div>
				<label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
				<input
					id="email"
					name="email"
					type="email"
					required
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
					placeholder="you@example.com"
				/>
			</div>
			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
				<input
					id="password"
					name="password"
					type="password"
					required
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
					placeholder="••••••••"
				/>
			</div>
			<button
				type="submit"
				class="w-full rounded-lg bg-emerald-600 text-white py-2 text-sm font-semibold hover:bg-emerald-700 transition-colors"
			>
				Sign In
			</button>
		</form>

		<p class="mt-4 text-center text-sm text-gray-500">
			No account? <a href="/register" class="text-emerald-600 hover:underline">Register</a>
		</p>
	</div>
</div>
