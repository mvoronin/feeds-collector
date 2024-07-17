<script>
  import { onMount } from 'svelte';
  import { channels, selectedChannel, items, selectedItem } from '$lib/stores';
  import { fetchChannels, fetchChannelItems, addChannel, deleteChannel } from '$lib/api';
  import { goto } from '$app/navigation';

  let newChannelName = '';
  let newChannelLink = '';

  onMount(async () => {
    const fetchedChannels = await fetchChannels();
    channels.set(fetchedChannels);
  });

  async function handleAddChannel() {
    const channel = await addChannel({ name: newChannelName, link: newChannelLink });
    channels.update(channels => [...channels, channel]);
    newChannelName = '';
    newChannelLink = '';
  }

  async function handleSelectChannel(channel) {
    const channelItems = await fetchChannelItems(channel.id);
    selectedChannel.set(channel);
    items.set(channelItems);
  }

  function handleSelectItem(item) {
    selectedItem.set(item);
  }
</script>

<div id="sidebar">
  <div class="logo">
    <img src="logo.png" alt="Logo">
    <span>Feeds Collector</span>
  </div>
  <div class="menu">
    <div class="menu-item" on:click="{handleAddChannel}">Button 1</div>
    <div class="menu-item">Button 2</div>
    <div class="menu-item">Button 3</div>
    {#each $channels as channel}
      <div class="menu-item" on:click={() => handleSelectChannel(channel)}>
        {channel.host}
      </div>
    {/each}
  </div>
</div>
<div id="content">
  <div id="feed-item-list">
    {#each $items as item}
      <div class="feed-item" on:click={() => handleSelectItem(item)}>
        {item.title}
      </div>
    {/each}
  </div>
  <div id="feed-item-content">
    {#if $selectedItem}
      <h2>{$selectedItem.title}</h2>
      <p>{$selectedItem.description}</p>
    {/if}
  </div>
</div>
