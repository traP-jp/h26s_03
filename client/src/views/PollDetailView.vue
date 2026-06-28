<template>
  <div class="page-container">
    <div class="header">
      <RouterLink to="/" class="back-link">＜ 戻る</RouterLink>
      <RouterLink
        :to="`/polls/${pollId}/input`"
        class="edit-link"
        v-if="me?.username && me?.username === poll?.created_by"
      >
        <EditIcon />
      </RouterLink>
      <h1>{{ poll?.name }}</h1>
    </div>
    <div v-if="poll" class="main-container">
      <div class="choices-container">
        <div class="select-container">
          <button
            type="button"
            class="choice-button"
            @click="selectChoice(1)"
            :class="{
              selected: myVote?.choice === 1,
              winner: poll?.result === 1,
              loser: hasResult && poll?.result !== 1,
            }"
            :disabled="isClosed || hasResult"
          >
            {{ poll?.choice1 }}
          </button>
          <TransitionGroup name="avatar-pop" tag="div" class="avatar-container">
            <img
              v-for="vote in visibleChoice1Votes"
              :key="vote.id"
              class="avatar"
              :class="{ expanded: isChoice1Expanded }"
              :data-name="vote.username"
              :title="vote.username"
              :src="`https://image-proxy.trap.jp/icon/${vote.username}?width=64&height=64`"
            />
            <button
              class="expand-button"
              @click="toggleChoice1Avatars"
              v-if="hiddenChoice1VoteCount > 0"
              key="choice1-expand-button"
            >
              {{ isChoice1Expanded ? "−" : `+${hiddenChoice1VoteCount}` }}
            </button>
          </TransitionGroup>
        </div>
        <div class="select-container">
          <button
            type="button"
            class="choice-button"
            @click="selectChoice(2)"
            :class="{
              selected: myVote?.choice === 2,
              winner: poll?.result === 2,
              loser: hasResult && poll?.result !== 2,
            }"
            :disabled="isClosed || hasResult"
          >
            {{ poll?.choice2 }}
          </button>
          <TransitionGroup name="avatar-pop" tag="div" class="avatar-container">
            <img
              v-for="vote in visibleChoice2Votes"
              :key="vote.id"
              :data-name="vote.username"
              :title="vote.username"
              :src="`https://image-proxy.trap.jp/icon/${vote.username}?width=64&height=64`"
              class="avatar"
              :class="{ expanded: isChoice2Expanded }"
            />
            <button
              class="expand-button"
              @click="toggleChoice2Avatars"
              v-if="hiddenChoice2VoteCount > 0"
              key="choice2-expand-button"
            >
              {{ isChoice2Expanded ? "−" : `+${hiddenChoice2VoteCount}` }}
            </button>
          </TransitionGroup>
        </div>
      </div>
      <div class="meta">
        <p>作成者: {{ poll?.created_by }}</p>
        <p v-if="poll?.due">期限: {{ new Date(poll.due).toLocaleString() }}</p>
        <p v-if="poll && !poll.due">期限: なし</p>
      </div>
      <p v-if="isClosed" class="closed-message">この投票は締め切られています。</p>
      <div v-if="hasResult" class="result-container">
        <p class="result-title">結果 : {{ winningChoiceName }} が勝利しました！</p>
        <p v-if="isCorrectVote === true" class="result-correct">的中しました！</p>
        <p v-else-if="isCorrectVote === false" class="result-incorrect">的中しませんでした…</p>
        <p v-else class="result-none">投票していません。</p>
      </div>
    </div>
    <div v-else-if="isLoading" class="main-container">
      <p>読み込み中...</p>
    </div>
    <div v-else class="main-container">
      <p>{{ errorMessage }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import EditIcon from "../components/EditIcon.vue";
import type { components } from "../gen/api-types";
import { createVote, deleteVote, getMe, getPoll, getVotes } from "../lib/api";

const route = useRoute();
const pollId = route.params.id as string;

type Poll = components["schemas"]["Poll"];
type Vote = components["schemas"]["Vote"];
type Me = components["schemas"]["Me"];
type VoteWebSocketMessage = components["schemas"]["VoteWebSocketMessage"];

const poll = ref<Poll | null>(null);
const votes = ref<Vote[]>([]);
const me = ref<Me | null>(null);
const isLoading = ref(false);
const errorMessage = ref("");
const wsConnection = ref<WebSocket | null>(null);

const collapsedAvatarLimit = 18;

const isChoice1Expanded = ref(false);
const isChoice2Expanded = ref(false);
const choice1Votes = computed(() => votes.value.filter((vote) => vote.choice === 1));
const choice2Votes = computed(() => votes.value.filter((vote) => vote.choice === 2));

const isClosed = computed(() => {
  if (!poll.value?.due) return false;
  return new Date(poll.value.due).getTime() <= Date.now();
});

const hasResult = computed(() => {
  return poll.value?.result === 1 || poll.value?.result === 2;
});

const winningChoiceName = computed(() => {
  if (!poll.value || !hasResult.value) return "";

  return poll.value.result === 1 ? poll.value.choice1 : poll.value.choice2;
});

const isCorrectVote = computed(() => {
  if (!hasResult.value || !myVote.value || !poll.value) return null;

  return myVote.value.choice === poll.value.result;
});

const myVote = computed(() => {
  if (!me.value) return null;

  return votes.value.find((vote) => vote.username === me.value!.username) ?? null;
});

const visibleChoice1Votes = computed(() => {
  if (isChoice1Expanded.value) {
    return choice1Votes.value;
  }

  return choice1Votes.value.slice(0, collapsedAvatarLimit);
});

const visibleChoice2Votes = computed(() => {
  if (isChoice2Expanded.value) {
    return choice2Votes.value;
  }

  return choice2Votes.value.slice(0, collapsedAvatarLimit);
});

const hiddenChoice1VoteCount = computed(() => {
  return Math.max(choice1Votes.value.length - collapsedAvatarLimit, 0);
});

const hiddenChoice2VoteCount = computed(() => {
  return Math.max(choice2Votes.value.length - collapsedAvatarLimit, 0);
});

const toggleChoice1Avatars = () => {
  isChoice1Expanded.value = !isChoice1Expanded.value;
};

const toggleChoice2Avatars = () => {
  isChoice2Expanded.value = !isChoice2Expanded.value;
};

const sendVoteWebSocketMessage = () => {
  if (!wsConnection.value) return;

  if (wsConnection.value.readyState !== WebSocket.OPEN) {
    console.warn("WebSocket is not open");
    return;
  }

  wsConnection.value.send(
    JSON.stringify({
      type: "vote",
      poll_id: String(pollId),
    }),
  );
};

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
  } catch (error) {
    console.error(error);
    errorMessage.value = "投票情報の取得に失敗しました。";
  } finally {
    isLoading.value = false;
  }
};

const selectChoice = async (choice: number) => {
  try {
    if (myVote.value && myVote.value.choice === choice) {
      return;
    }

    if (hasResult.value) {
      alert("この投票は結果が確定しています。");
      return;
    }

    if (isClosed.value) {
      alert("この投票は締め切られています。");
      return;
    }

    if (myVote.value) {
      await deleteVote(Number(pollId), myVote.value.id);
    }

    await createVote(Number(pollId), choice, 1); //最後の引数は仮のbet
    sendVoteWebSocketMessage();
    await fetchVoteList();
  } catch (error) {
    console.error("Error creating vote:", error);
    alert("投票に失敗しました。もう一度試してください。");
  }
};

const connectWebSocket = () => {
  try {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    // const wsUrl = `${protocol}//localhost:8080/api/ws?poll_id=${pollId}`;
    const wsUrl = `${protocol}//${window.location.host}/api/ws?poll_id=${pollId}`;
    wsConnection.value = new WebSocket(wsUrl);
    wsConnection.value.onmessage = (event: MessageEvent) => {
      try {
        console.log(event.data);
        const message: VoteWebSocketMessage = JSON.parse(event.data);
        if (message.type === "vote") {
          fetchVoteList();
        }
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    wsConnection.value.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    wsConnection.value.onclose = () => {
      console.log("WebSocket closed");
    };
  } catch (error) {
    console.error("Failed to connect WebSocket:", error);
  }
};

const disconnectWebSocket = () => {
  if (wsConnection.value) {
    wsConnection.value.close();
    wsConnection.value = null;
  }
};

onMounted(() => {
  fetchPageData();
  connectWebSocket();
});

onBeforeUnmount(() => {
  disconnectWebSocket();
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

.main-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  margin-top: 20px;
  padding: 20px;
}

.choices-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 48px;
  width: 100%;
  max-width: 760px;
  align-items: start;
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
.choice-button:hover:not(:disabled) {
  background-color: #374151;
}
.choice-button.selected {
  background-color: #101e40;
  color: #8ec5ff;
  border: 2px solid #51a2ff;
}
.choice-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
.choice-button.winner {
  background: #dbeafe;
  color: #1d4ed8;
  border-color: #93c5fd;
}

.choice-button.loser {
  opacity: 0.45;
}

.choice-button.selected.winner {
  box-shadow:
    0 0 0 3px rgba(59, 130, 246, 0.25),
    0 2px 8px rgba(0, 0, 0, 0.06);
}

.choice-button.selected.loser {
  background: #1f2937;
  color: #9ca3af;
  border-color: #4b5563;
  opacity: 0.75;
}

.result-container {
  margin: 24px auto 0;
  max-width: 360px;
  padding: 16px;
  border-radius: 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
}

.result-title {
  margin: 0 0 8px;
  font-size: 14px;
  color: #6b7280;
  font-weight: 700;
}

.result-choice {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  color: #166534;
}

.result-correct {
  margin: 12px 0 0;
  color: #16a34a;
  font-weight: 700;
}

.result-incorrect {
  margin: 12px 0 0;
  color: #dc2626;
  font-weight: 700;
}

.result-none {
  margin: 12px 0 0;
  color: #6b7280;
}
.choice-button.selected:disabled {
  opacity: 0.8;
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
  flex-wrap: wrap;
  max-width: 320px;
  min-height: 44px;
  margin: 10px auto 0;
}
.avatar-container.expanded {
  max-width: 320px;
}
.expand-button {
  min-width: 30px;
  height: 30px;
  padding: 0 10px;
  border-radius: 999px;
  display: grid;
  place-items: center;
  background: #334155;
  color: #ffffff;
  font-size: 13px;
  font-weight: 700;
  border: 2px solid rgba(255, 255, 255, 0.35);
  cursor: pointer;
}
.expand-button:hover {
  background: #475569;
}
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  margin: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #ffffff;
  border: 2px solid rgba(255, 255, 255, 0.75);
}

.avatar-pop-enter-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.avatar-pop-enter-from {
  opacity: 0;
  transform: scale(0.4) translateY(8px);
}

.avatar-pop-enter-to {
  opacity: 1;
  transform: scale(1) translateY(0);
}

.avatar-pop-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.avatar-pop-leave-from {
  opacity: 1;
  transform: scale(1);
}

.avatar-pop-leave-to {
  opacity: 0;
  transform: scale(0.4);
}

.avatar-pop-move {
  transition: transform 0.25s ease;
}
.meta {
  margin-top: 32px;
  font-size: 13px;
  color: #6b7280;
  display: flex;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}
.closed-message {
  margin-top: 20px;
  color: #6b7280;
  font-size: 14px;
  font-weight: 600;
}

@media (max-width: 700px) {
  .choices-container {
    grid-template-columns: 1fr;
    gap: 28px;
    max-width: 360px;
  }

  .choice-button {
    max-width: 280px;
  }

  .avatar-container {
    max-width: 280px;
  }
}
</style>
