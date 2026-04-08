<script lang="ts">
	interface Props {
		open: boolean;
		title?: string;
		message?: string;
		confirmLabel?: string;
		onconfirm: () => void;
		oncancel: () => void;
	}

	let {
		open,
		title = 'Are you sure?',
		message = 'This action cannot be undone.',
		confirmLabel = 'Delete',
		onconfirm,
		oncancel
	}: Props = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();

	$effect(() => {
		if (open && dialogEl) {
			dialogEl.querySelector<HTMLElement>('[data-autofocus]')?.focus();
		}
	});
</script>

<svelte:window
	onkeydown={(e) => {
		if (open && e.key === 'Escape') oncancel();
	}}
/>

{#if open}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={(e) => { if (e.target === e.currentTarget) oncancel(); }}
	>
		<div
			bind:this={dialogEl}
			role="alertdialog"
			aria-modal="true"
			aria-labelledby="confirm-title"
			aria-describedby="confirm-message"
			class="bg-white rounded-2xl border border-gray-200 shadow-lg w-full max-w-sm mx-4 p-6"
		>
			<h2 id="confirm-title" class="text-base font-semibold text-gray-900 mb-2">{title}</h2>
			<p id="confirm-message" class="text-sm text-gray-500 mb-6">{message}</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={oncancel}
					class="rounded-lg border border-gray-200 text-gray-600 text-sm px-4 py-2 hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					data-autofocus
					onclick={onconfirm}
					class="rounded-lg bg-red-600 text-white text-sm px-4 py-2 font-semibold hover:bg-red-700 transition-colors"
				>
					{confirmLabel}
				</button>
			</div>
		</div>
	</div>
{/if}
