<template>
  <div class="page-container">
    <div v-if="isResultAnnouncing" class="result-announcement-overlay">
      <div class="result-announcement-card">
        <p class="announcement-label">結果発表！</p>
        <p class="announcement-winner">{{ winningChoiceName }} が勝利！</p>

        <p v-if="isCorrectVote === true" class="announcement-correct">的中しました！</p>
        <p v-else-if="isCorrectVote === false" class="announcement-incorrect">
          的中しませんでした…
        </p>
        <p v-else class="announcement-none">投票していません</p>
      </div>
    </div>
    <div class="header">
      <RouterLink to="/" class="back-link"> < 戻る</RouterLink>
      <div class="header-icons">
        <a :href="shareUrl" target="_blank" rel="noopener noreferrer" class="share-link">
          <ShareIcon />
        </a>
        <RouterLink
          :to="`/polls/${pollId}/input`"
          class="edit-link"
          v-if="me && me.username === poll?.created_by"
        >
          <EditIcon />
        </RouterLink>
      </div>
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
      <div class="reaction-buttons">
        <button
          v-for="reaction in reactions"
          @click="sendReaction(reaction)"
          :key="reaction"
          class="reaction-button"
          type="button"
          :aria-label="`Send ${reaction} reaction`"
          :title="reaction"
        >
          <img :src="reactionIcons[reaction]" :alt="reaction" width="32" height="32" />
        </button>
      </div>
      <div class="reaction-container" ref="reaction-container"></div>
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
import confetti from "@hiseb/confetti";
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef } from "vue";
import { useRoute } from "vue-router";

import fire_icon from "../assets/reactions/fire.svg";
import heart_icon from "../assets/reactions/heart.svg";
import moneybag_icon from "../assets/reactions/moneybag.svg";
import open_mouth_icon from "../assets/reactions/open_mouth.svg";
import star_struck_icon from "../assets/reactions/star_struck.svg";
import thumbs_up_icon from "../assets/reactions/thumbs_up.svg";
import zany_face_icon from "../assets/reactions/zany_face.svg";
import EditIcon from "../components/EditIcon.vue";
import ShareIcon from "../components/ShareIcon.vue";
import type { components } from "../gen/api-types";
import { createVote, deleteVote, getMe, getPoll, getVotes } from "../lib/api";

const reactionContainerRef = useTemplateRef("reaction-container");

const reactions = [
  "fire",
  "thumbs_up",
  "heart",
  "open_mouth",
  "zany_face",
  "star_struck",
  "moneybag",
];

const reactionIcons: Record<string, string> = {
  fire: fire_icon,
  thumbs_up: thumbs_up_icon,
  heart: heart_icon,
  open_mouth: open_mouth_icon,
  zany_face: zany_face_icon,
  star_struck: star_struck_icon,
  moneybag: moneybag_icon,
};

const route = useRoute();
const pollId = route.params.id as string;

type Poll = components["schemas"]["Poll"];
type Vote = components["schemas"]["Vote"];
type Me = components["schemas"]["Me"];
type WebSocketMessage = components["schemas"]["WebSocketMessage"];

const poll = ref<Poll | null>(null);
const votes = ref<Vote[]>([]);
const me = ref<Me | null>(null);
const isLoading = ref(false);
const errorMessage = ref("");
const wsConnection = ref<WebSocket | null>(null);
const isResultAnnouncing = ref(false);
const announcementTimer = ref<number | null>(null);

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

const sendReaction = (reaction: string) => {
  if (!wsConnection.value) return;

  if (wsConnection.value.readyState !== WebSocket.OPEN) {
    console.warn("WebSocket is not open");
    return;
  }

  wsConnection.value.send(
    JSON.stringify({
      type: "reaction",
      poll_id: String(pollId),
      reaction: reaction,
    }),
  );
};

const showReaction = (reaction: string) => {
  if (!reactionContainerRef.value) return;
  const reactionIcon = reactionIcons[reaction];
  if (!reactionIcon) return;

  const reactionElement = document.createElement("img");
  reactionElement.src = reactionIcon;
  reactionElement.alt = reaction;
  reactionElement.className = "reaction-animation";
  const rotation = Math.random() * 28 - 14;
  const drift = Math.random() * 56 - 28;
  reactionElement.style.setProperty("--reaction-x", `${Math.random() * 72 + 14}%`);
  reactionElement.style.setProperty("--reaction-drift-mid", `${drift * 0.45}px`);
  reactionElement.style.setProperty("--reaction-drift-end", `${drift}px`);
  reactionElement.style.setProperty("--reaction-rotate-start", `${rotation * -0.45}deg`);
  reactionElement.style.setProperty("--reaction-rotate-mid", `${rotation * 0.45}deg`);
  reactionElement.style.setProperty("--reaction-rotate-end", `${rotation}deg`);
  reactionElement.style.setProperty("--reaction-delay", `${Math.random() * 0.12}s`);
  reactionContainerRef.value.appendChild(reactionElement);

  setTimeout(() => {
    if (reactionContainerRef.value && reactionElement.parentNode === reactionContainerRef.value) {
      reactionContainerRef.value.removeChild(reactionElement);
    }
  }, 2000);
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

    await createVote(Number(pollId), choice, 0);
    sendVoteWebSocketMessage();
    await fetchVoteList();
  } catch (error) {
    console.error("Error creating vote:", error);
    alert("投票に失敗しました。もう一度試してください。");
  }
};

const runConfetti = () => {
  confetti({
    position: {
      x: window.innerWidth * 0.5,
      y: window.innerHeight * 0.3,
    }, // Origin position
    count: 100, // Number of particles
    size: 1, // Size of the particles
    velocity: 200, // Initial particle velocity
    fade: false, // Particles fall off the screen, or fade out
  });

  setTimeout(() => {
    confetti({
      position: { x: window.innerWidth * 0.5, y: window.innerHeight * 0.3 }, // Origin position
      count: 100, // Number of particles
      size: 1, // Size of the particles
      velocity: 200, // Initial particle velocity
      fade: false, // Particles fall off the screen, or fade out
    });
  }, 300);

  setTimeout(() => {
    confetti({
      position: { x: window.innerWidth * 0.5, y: window.innerHeight * 0.3 }, // Origin position
      count: 100, // Number of particles
      size: 1, // Size of the particles
      velocity: 200, // Initial particle velocity
      fade: false, // Particles fall off the screen, or fade out
    });
  }, 600);
};

const showResultAnnouncement = () => {
  isResultAnnouncing.value = true;
  runConfetti();

  if (announcementTimer.value !== null) {
    window.clearTimeout(announcementTimer.value);
  }

  announcementTimer.value = window.setTimeout(() => {
    isResultAnnouncing.value = false;
    announcementTimer.value = null;
  }, 2200);
};

const connectWebSocket = () => {
  try {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/ws?poll_id=${pollId}`;
    wsConnection.value = new WebSocket(wsUrl);
    wsConnection.value.onmessage = (event: MessageEvent) => {
      try {
        console.log(event.data);
        const message: WebSocketMessage = JSON.parse(event.data);
        if (message.type === "vote") {
          fetchVoteList();
        }
        if (message.type === "poll_status") {
          if (poll.value && message?.poll_id === Number(pollId)) {
            poll.value = {
              ...poll.value,
              result: message.result,
            };
            fetchPageData();
            if (message.result === 1 || message.result === 2) {
              showResultAnnouncement();
            }
          }
        }
        if (message.type === "reaction") {
          showReaction(message.reaction);
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

  if (announcementTimer.value !== null) {
    window.clearTimeout(announcementTimer.value);
    announcementTimer.value = null;
  }
};

onMounted(() => {
  fetchPageData();
  connectWebSocket();
});

onBeforeUnmount(() => {
  disconnectWebSocket();
});
//共有機能
const shareUrl = computed(() => {
  if (me.value === null || poll.value === null) {
    return "";
  }

  let text = "";

  if (hasResult.value) {
    text = `投票「${poll.value.name}」の結果が出ました！\n` + `勝者: ${winningChoiceName.value}\n`;

    if (isCorrectVote.value === true) {
      text += `${me.value.username}は的中しました！\n`;
    } else if (isCorrectVote.value === false) {
      text += `${me.value.username}は外れました…\n`;
    } else {
      text += "投票は締め切られました\n";
    }

    text += window.location.href;
  } else {
    text = `投票「${poll.value.name}」に参加してください！\n` + `${window.location.href}`;
  }

  return `https://q.trap.jp/share-target?text=${encodeURIComponent(text)}`;
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

.header-icons {
  align-self: flex-end;
  display: flex;
  align-items: center;
  gap: 8px;
}

.back-link {
  align-self: flex-start;
  margin: 10px;
  text-decoration: none;
  color: #ffffff;
}

.edit-link {
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
  min-height: 50px;
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
  border-color: hsl(215, 14%, 34%);
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

.reaction-buttons {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
  width: min(100%, 460px);
  margin: 30px auto 12px;
  padding: 10px;
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.72);
}

.reaction-button {
  display: grid;
  place-items: center;
  width: 46px;
  height: 46px;
  padding: 0;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background-color: rgba(15, 23, 42, 0.72);
  border-radius: 50%;
  cursor: pointer;
}

.reaction-button:hover {
  border-color: rgba(142, 197, 255, 0.7);
  background-color: rgba(51, 65, 85, 0.98);
}

.reaction-button:active {
  transform: scale(0.94);
}

.reaction-button:focus-visible {
  outline: 3px solid rgba(142, 197, 255, 0.85);
  outline-offset: 3px;
}

.reaction-button img {
  width: 30px;
  height: 30px;
  filter: drop-shadow(0 3px 5px rgba(0, 0, 0, 0.25));
  pointer-events: none;
}

.reaction-container {
  position: fixed;
  left: 50%;
  bottom: 84px;
  z-index: 1200;
  width: min(520px, 92vw);
  height: min(460px, 58vh);
  pointer-events: none;
  transform: translateX(-50%);
  overflow: visible;
}

.reaction-container :deep(.reaction-animation) {
  position: absolute;
  left: var(--reaction-x, 50%);
  bottom: 0;
  width: 48px;
  height: 48px;
  object-fit: contain;
  filter: drop-shadow(0 12px 16px rgba(0, 0, 0, 0.28));
  transform: translateX(-50%);
  animation: reaction-float 2s var(--reaction-delay, 0s) cubic-bezier(0.18, 0.84, 0.24, 1) forwards;
  will-change: opacity, transform;
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
  margin-top: 0px;
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

  .reaction-buttons {
    gap: 8px;
    max-width: 340px;
    padding: 8px;
    border-radius: 24px;
  }

  .reaction-button {
    width: 42px;
    height: 42px;
  }

  .reaction-button img {
    width: 28px;
    height: 28px;
  }

  .reaction-container {
    bottom: 68px;
    height: min(420px, 54vh);
  }
}
.result-announcement-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;

  background: rgba(15, 23, 43, 0.72);
  backdrop-filter: blur(6px);

  animation: overlay-fade-in 0.2s ease-out;
}

.result-announcement-card {
  width: min(420px, 92vw);
  padding: 32px 30px 28px;
  border-radius: 20px;

  background: #f8fafc;
  color: #0f172a;

  border: 1px solid #e2e8f0;

  box-shadow:
    0 24px 70px rgba(0, 0, 0, 0.35),
    0 2px 8px rgba(15, 23, 42, 0.12);

  text-align: center;
  animation: result-card-pop 0.42s cubic-bezier(0.16, 1, 0.3, 1);
}

.announcement-label {
  margin: 0 0 10px;
  color: #334155;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.announcement-winner {
  margin: 0;
  color: #0f172a;
  font-size: clamp(34px, 7vw, 48px);
  font-weight: 900;
  line-height: 1.08;
}

.announcement-correct {
  color: #16a34a;
}

.announcement-incorrect {
  color: #dc2626;
}

.announcement-none {
  color: #64748b;
}

.result-announcement-card.correct {
  border-top: 6px solid #22c55e;
}

.result-announcement-card.incorrect {
  border-top: 6px solid #ef4444;
}

.result-announcement-card.none {
  border-top: 6px solid #94a3b8;
}

@keyframes overlay-fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes result-card-pop {
  0% {
    opacity: 0;
    transform: scale(0.88) translateY(18px);
  }

  70% {
    opacity: 1;
    transform: scale(1.03) translateY(0);
  }

  100% {
    transform: scale(1);
  }
}

@keyframes reaction-float {
  0% {
    opacity: 0;
    transform: translate(-50%, 30px) scale(0.55) rotate(var(--reaction-rotate-start, 0deg));
  }

  12% {
    opacity: 1;
    transform: translate(-50%, 0) scale(1.1) rotate(0deg);
  }

  55% {
    opacity: 1;
    transform: translate(calc(-50% + var(--reaction-drift-mid, 0px)), -180px) scale(1)
      rotate(var(--reaction-rotate-mid, 0deg));
  }

  100% {
    opacity: 0;
    transform: translate(calc(-50% + var(--reaction-drift-end, 0px)), -320px) scale(0.72)
      rotate(var(--reaction-rotate-end, 0deg));
  }
}
</style>
