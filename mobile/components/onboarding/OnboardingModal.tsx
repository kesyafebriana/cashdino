import { useState, useEffect } from "react";
import { View, Text, Modal, Pressable } from "react-native";
import { useRouter } from "expo-router";
import AsyncStorage from "@react-native-async-storage/async-storage";

import { StepIntro } from "./StepIntro";
import { StepHowItWorks } from "./StepHowItWorks";
import { StepPrivacy } from "./StepPrivacy";
import { useUser } from "@/contexts/UserContext";

const APP_VERSION = require("../../app.json").expo.version;
const STORAGE_KEY_PREFIX =
  process.env.EXPO_PUBLIC_STORAGE_KEY_ONBOARDING || "onboarding_shown_version";
const TOTAL_STEPS = 3;

export function OnboardingModal() {
  const router = useRouter();
  const { currentUser } = useUser();
  const [visible, setVisible] = useState(false);
  const [step, setStep] = useState(0);

  const storageKey = currentUser
    ? `${STORAGE_KEY_PREFIX}:${currentUser.id}`
    : null;

  useEffect(() => {
    if (!storageKey) return;
    setStep(0);
    AsyncStorage.getItem(storageKey).then((shownVersion) => {
      if (shownVersion !== APP_VERSION) {
        setVisible(true);
      }
    });
  }, [storageKey]);

  const dismiss = async () => {
    setVisible(false);
    setStep(0);
    if (storageKey) {
      await AsyncStorage.setItem(storageKey, APP_VERSION);
    }
  };

  const handleNext = () => {
    if (step < TOTAL_STEPS - 1) {
      setStep(step + 1);
    } else {
      dismiss();
      router.push("/leaderboard");
    }
  };

  const steps = [<StepIntro key={0} />, <StepHowItWorks key={1} />, <StepPrivacy key={2} />];

  const primaryLabel =
    step < TOTAL_STEPS - 1 ? "Next" : "View leaderboard →";
  const secondaryLabel = step < TOTAL_STEPS - 1 ? "Skip" : "Got it";

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      statusBarTranslucent
    >
      <View
        className="flex-1 items-center justify-center"
        style={{ backgroundColor: "rgba(0,0,0,0.5)", padding: 24 }}
      >
        <View className="w-full bg-white overflow-hidden" style={{ borderRadius: 24 }}>
          {/* Step content */}
          {steps[step]}

          {/* Dots + buttons */}
          <View style={{ paddingHorizontal: 24, paddingTop: 20, paddingBottom: 28 }}>
            {/* Page dots */}
            <View
              className="flex-row items-center justify-center"
              style={{ gap: 6, marginBottom: 20 }}
            >
              {Array.from({ length: TOTAL_STEPS }).map((_, i) => (
                <View
                  key={i}
                  className="rounded-full"
                  style={{
                    width: 8,
                    height: 8,
                    backgroundColor: i === step ? "#e65100" : "#ddd",
                  }}
                />
              ))}
            </View>

            {/* Primary button */}
            <Pressable
              onPress={handleNext}
              className="items-center"
              style={{
                backgroundColor: "#e65100",
                borderRadius: 14,
                padding: 14,
                marginBottom: 8,
              }}
            >
              <Text className="text-white font-semibold" style={{ fontSize: 15 }}>
                {primaryLabel}
              </Text>
            </Pressable>

            {/* Secondary button */}
            <Pressable onPress={dismiss} className="items-center" style={{ padding: 8 }}>
              <Text style={{ fontSize: 13, color: "#999" }}>{secondaryLabel}</Text>
            </Pressable>
          </View>
        </View>
      </View>
    </Modal>
  );
}
