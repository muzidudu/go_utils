<script lang="ts">
	import { cn } from '$lib/utils';

	interface Props {
		class?: string;
		type?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean;
	}

	let {
		class: className,
		type = 'text',
		value = $bindable(''),
		placeholder,
		disabled = false
	}: Props = $props();
</script>

<input
	{type}
	{placeholder}
	{disabled}
	bind:value
	class={cn(
		'flex h-9 w-full rounded-md border border-slate-200 bg-white px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-slate-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
/>
