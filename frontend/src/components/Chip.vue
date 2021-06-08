<template>
  <div class="px-4 h-8 flex justify-center items-center rounded-full text-xs" :class="classObject">
    <slot></slot>
  </div> 
</template>

<script>
export default {
  name: 'Chip',
  props: {
    focused: Boolean,
  },
  computed: {
    classObject: function() {
      return {
        'text-secondary': !this.focused,
        'bg-purple-100': !this.focused,
        'text-white': this.focused,
        'bg-primary': this.focused,
      }
    }
  }
}
</script>

<style>

</style>
