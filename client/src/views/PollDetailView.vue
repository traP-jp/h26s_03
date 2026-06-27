<template>
  <div class="page-container">
    <div class="header">
      <Router-link to="/" class="back-link">＜ 戻る</Router-link>
      <RouterLink
        :to="`/polls/${pollId}/input`"
        class="edit-link"
        v-if="me?.username === poll?.created_by"
      >
        <img :src="editIcon" alt="編集" class="edit-icon" />
      </RouterLink>
      <h1>{{ poll?.name }}</h1>
    </div>
    <div class="main-container">
      <div class="select-container">
        <button
          type="button"
          class="choice-button"
          @click="selectChoice(1)"
          :class="{ selected: selectedChoice === 1 }"
        >
          {{ poll?.choice1 }}
        </button>
        <div class="avatar-container">
          <img
            v-for="vote in choice1Votes"
            :key="vote.id"
            class="avatar"
            :data-name="vote.username"
            :title="vote.username"
            :src="`https://image-proxy.trap.jp/icon/${vote.username}?width=64&height=64`"
          />
        </div>
      </div>
      <div class="select-container">
        <button
          type="button"
          class="choice-button"
          @click="selectChoice(2)"
          :class="{ selected: selectedChoice === 2 }"
        >
          {{ poll?.choice2 }}
        </button>
        <div class="avatar-container">
          <img
            v-for="vote in choice2Votes"
            :key="vote.id"
            :data-name="vote.username"
            :title="vote.username"
            :src="`https://image-proxy.trap.jp/icon/${vote.username}?width=64&height=64`"
            class="avatar"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import editIcon from "../assets/editIcon.svg";
import type { components } from "../gen/api-types";
import { createVote, deleteVote, getMe, getPoll, getVotes } from "../lib/api";

const route = useRoute();
const pollId = route.params.id as string;

type Poll = components["schemas"]["Poll"];
type Vote = components["schemas"]["Vote"];
type Me = components["schemas"]["Me"];

const poll = ref<Poll | null>(null);
const votes = ref<Vote[]>([]);
const selectedChoice = ref<number | null>(null);
const me = ref<Me | null>(null);
const isLoading = ref(false);
const errorMessage = ref("");

const choice1Votes = computed(() => votes.value.filter((vote) => vote.choice === 1));
const choice2Votes = computed(() => votes.value.filter((vote) => vote.choice === 2));

const myVote = computed(() => {
  if (!me.value) return null;

  return votes.value.find((vote) => vote.username === me.value!.username) ?? null;
});

const fetchVoteList = async () => {
  const data = await getVotes(Number(pollId));
  votes.value = data;
};

const fetchPageData = async () => {
  isLoading.value = true;
  errorMessage.value = "";

  try {
    const [pollData, voteData, meData] = await Promise.all([
      getPoll(Number(pollId)),
      getVotes(Number(pollId)),
      getMe(),
    ]);

    poll.value = pollData;
    votes.value = voteData;
    me.value = meData;
    if (myVote.value) {
      selectedChoice.value = myVote.value.choice;
    }
  } catch (error) {
    console.error(error);
    errorMessage.value = "投票情報の取得に失敗しました。";
  } finally {
    isLoading.value = false;
  }
};

const selectChoice = async (choice: number) => {
  selectedChoice.value = choice;
  try {
    if (myVote.value && myVote.value.choice === choice) {
      return;
    }

    if (myVote.value) {
      await deleteVote(Number(pollId), myVote.value.id);
    }

    await createVote(Number(pollId), choice, 1); //最後の引数は仮のbet
    await fetchVoteList();
  } catch (error) {
    console.error("Error creating vote:", error);
  }
};

onMounted(() => {
  fetchPageData();
});
</script>

<style scoped>
.page-container {
  text-align: center;
  background-color: #0f172b;
  color: #ffffff;
  min-height: 100vh;
}
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  margin-bottom: 20px;
}
.back-link {
  align-self: flex-start;
  margin: 10px;
  text-decoration: none;
  color: #ffffff;
}
.edit-link {
  align-self: flex-end;
  margin: 10px;
  text-decoration: none;
  color: #ffffff;
}

.edit-icon {
  width: 24px;
  height: 24px;
}

.main-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-top: 20px;
  padding: 20px;
}
.select-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}
.choice-button {
  width: 100%;
  max-width: 280px;
  min-height: 72px;
  font-size: 20px;
  font-weight: 700;
  margin: 10px;
  cursor: pointer;
  padding: 12px 24px;
  border: 2px solid #e5e7eb;
  background: #0f172b;
  color: #ffffff;
}
.choice-button:hover {
  background-color: #374151;
}
.choice-button.selected {
  background-color: #101e40;
  color: #8ec5ff;
  border: 2px solid #51a2ff;
}
.icon-container {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 10px;
}
.avatar-container {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 10px;
}
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  margin: 0 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #ffffff;
}
</style>
