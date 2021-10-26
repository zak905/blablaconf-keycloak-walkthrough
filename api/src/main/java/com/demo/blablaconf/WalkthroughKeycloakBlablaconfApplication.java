package com.demo.blablaconf;

import java.util.List;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@SpringBootApplication
public class WalkthroughKeycloakBlablaconfApplication {

	public static void main(String[] args) {
		SpringApplication.run(WalkthroughKeycloakBlablaconfApplication.class, args);
	}

}

@RestController
class ConferenceListController {
	@GetMapping
	public List<String> listConferences() {
		return List.of("A Walkthrough Keycloak", "The right questions to ask when using microservices", "Distributed systems 102", "Binary Exploitation 101", "Software Quality Assurance Craftsmanship", "Building Microservices Architecture The Right Way", "want to talk about what I do the most holding IT events");
	}
}
