import { defineStore } from 'pinia'
import { ref } from 'vue'
import {NewServer} from "../types/server";

export const useServerStore = defineStore('user', () => {
    const server = ref({...NewServer})
    return { server }
})