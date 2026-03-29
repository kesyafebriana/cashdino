import { View, Text, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useUser } from "@/contexts/UserContext";
import { Skeleton } from "@/components/Skeleton";

function ProfileSkeleton() {
  return (
    <View className="flex-1 bg-white px-6 pt-6">
      <View className="items-center mb-8">
        <Skeleton width={80} height={80} borderRadius={40} />
        <View style={{ height: 12 }} />
        <Skeleton width={120} height={20} />
        <View style={{ height: 8 }} />
        <Skeleton width={180} height={14} />
      </View>

      <Skeleton width={140} height={18} style={{ marginBottom: 8 }} />
      <Skeleton width={260} height={12} style={{ marginBottom: 16 }} />

      <View className="gap-2">
        {[1, 2, 3].map((i) => (
          <View
            key={i}
            className="flex-row items-center rounded-2xl px-4 py-3 gap-3"
            style={{ backgroundColor: "#f5f5f5" }}
          >
            <Skeleton width={40} height={40} borderRadius={20} />
            <View>
              <Skeleton width={100} height={14} style={{ marginBottom: 4 }} />
              <Skeleton width={150} height={12} />
            </View>
          </View>
        ))}
      </View>
    </View>
  );
}

export default function ProfileScreen() {
  const { currentUser, users, switchUser, loading, error, retry } = useUser();

  if (loading) {
    return <ProfileSkeleton />;
  }

  if (error) {
    return (
      <View className="flex-1 bg-white items-center justify-center px-6">
        <Ionicons name="cloud-offline-outline" size={48} color="#ccc" />
        <Text className="text-base text-gray-500 mt-3 mb-1">
          Could not load users
        </Text>
        <Text className="text-xs text-gray-400 mb-4 text-center">
          Make sure the backend is running and EXPO_PUBLIC_API_URL is set
          correctly in .env
        </Text>
        <Pressable
          onPress={retry}
          className="bg-gray-900 rounded-xl px-6 py-2.5"
        >
          <Text className="text-sm font-semibold text-white">Retry</Text>
        </Pressable>
      </View>
    );
  }

  return (
    <View className="flex-1 bg-white px-6 pt-6">
      {/* Avatar + current user */}
      <View className="items-center mb-8">
        <View className="w-20 h-20 rounded-full bg-gray-100 items-center justify-center mb-3">
          <Ionicons name="person" size={40} color="#999" />
        </View>
        {currentUser ? (
          <>
            <Text className="text-xl font-bold text-gray-900">
              {currentUser.username}
            </Text>
            <Text className="text-sm text-gray-500">{currentUser.email}</Text>
          </>
        ) : (
          <Text className="text-sm text-gray-400">No user selected</Text>
        )}
      </View>

      {/* Switch account */}
      <Text className="text-base font-bold text-gray-900 mb-3">
        Switch account
      </Text>
      <Text className="text-xs text-gray-400 mb-4">
        Tap a username to simulate logging in as that user
      </Text>

      <View className="gap-2">
        {users.map((user) => {
          const isActive = user.id === currentUser?.id;
          return (
            <Pressable
              key={user.id}
              onPress={() => switchUser(user)}
              className="flex-row items-center justify-between rounded-2xl px-4 py-3"
              style={{
                backgroundColor: isActive ? "#1a1a1a" : "#f5f5f5",
              }}
            >
              <View className="flex-row items-center gap-3">
                <View
                  className="w-10 h-10 rounded-full items-center justify-center"
                  style={{
                    backgroundColor: isActive ? "#333" : "#e0e0e0",
                  }}
                >
                  <Ionicons
                    name="person"
                    size={20}
                    color={isActive ? "#fff" : "#999"}
                  />
                </View>
                <View>
                  <Text
                    className="text-sm font-semibold"
                    style={{ color: isActive ? "#fff" : "#1a1a1a" }}
                  >
                    {user.username}
                  </Text>
                  <Text
                    className="text-xs"
                    style={{ color: isActive ? "#aaa" : "#888" }}
                  >
                    {user.email}
                  </Text>
                </View>
              </View>
              {isActive && (
                <View className="bg-green-500 rounded-full px-2 py-0.5">
                  <Text className="text-xs font-semibold text-white">
                    Active
                  </Text>
                </View>
              )}
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}
