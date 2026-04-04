#import <Foundation/Foundation.h>
#import <UserNotifications/UserNotifications.h>

// sendUserNotification requests authorization (once, silently after first grant)
// then posts a UNUserNotificationCenter notification. Must be run from within a
// proper .app bundle — plain binaries lack the bundle proxy LaunchServices needs.
// Dispatched to the main queue because UNUserNotificationCenter is not thread-safe.
void sendUserNotification(const char *title, const char *subtitle) {
    NSString *ttl = [NSString stringWithUTF8String:title];
    NSString *sub = subtitle && strlen(subtitle) > 0
        ? [NSString stringWithUTF8String:subtitle]
        : nil;

    dispatch_async(dispatch_get_main_queue(), ^{
        UNUserNotificationCenter *center =
            [UNUserNotificationCenter currentNotificationCenter];

        [center getNotificationSettingsWithCompletionHandler:^(UNNotificationSettings *settings) {
            NSLog(@"gl1tch-notify: auth status = %ld", (long)settings.authorizationStatus);
        }];

        [center requestAuthorizationWithOptions:(UNAuthorizationOptionAlert |
                                                 UNAuthorizationOptionSound)
                              completionHandler:^(BOOL granted, NSError *error) {
            NSLog(@"gl1tch-notify: granted=%d error=%@", granted, error);
            if (!granted) { return; }

            UNMutableNotificationContent *content =
                [[UNMutableNotificationContent alloc] init];
            content.title = @"gl1tch";
            if (sub) {
                content.subtitle = sub;
                content.body = ttl;
            } else {
                content.body = ttl;
            }

            NSLog(@"gl1tch-notify: posting — title=%@ subtitle=%@",
                  content.title, content.subtitle);

            UNNotificationRequest *req = [UNNotificationRequest
                requestWithIdentifier:[[NSUUID UUID] UUIDString]
                              content:content
                              trigger:nil];

            [center addNotificationRequest:req withCompletionHandler:^(NSError *e) {
                NSLog(@"gl1tch-notify: addNotificationRequest done, err=%@", e);
            }];
        }];
    });
}
