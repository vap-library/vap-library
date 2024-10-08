# Given a yaml config file in the prescribed format, generate the required files which can be used to apply the chosen policies on a cluster.
# One file will be generated containing the policies, one for the bindings (these will be created from scratch as per the given yaml), and one for any CRDs (not all policies require these).
# If a policy file begins with '---' as the opening line, this will be stripped.

import yaml
import os
import sys

class Release:
    def __init__(self, config_path, policies_domain, policies_output_filename, bindings_output_filename, crds_output_filename, kustomization_output_filename):
        self.__policies_domain = policies_domain
        self.__release_output_dir = "release"
        self.__policies_file = policies_output_filename
        self.__bindings_file = bindings_output_filename
        self.__crds_file = crds_output_filename
        self.__kustomization_file = kustomization_output_filename
        self.__config = self.__parse_config(config_path)

    def __parse_config(self, config_path):
        with open(config_path, 'r') as file:
          config = yaml.safe_load(file)
        return config
    
    def __append_policy(self, subdirectory):
        policy_path = os.path.join('../policies', subdirectory, 'policy.yaml')
        if os.path.exists(policy_path):
            with open(policy_path, 'r') as file:
                policy_content = file.read().rstrip('\n').lstrip('---\n')
  
            with open(os.path.join(self.__release_output_dir, self.__policies_file), 'a') as file:
                file.write(policy_content + "\n")
                file.write('---\n')
        else:
          print(f"Warning: {policy_path} does not exist.")

    def __create_binding(self, policy_name, binding_name, param_ref=None, validation_actions=None, match_resources=None):
        binding = {
            'apiVersion': 'admissionregistration.k8s.io/v1',
            'kind': 'ValidatingAdmissionPolicyBinding',
            'metadata': {
                'name': binding_name
            },
            'spec': {
                'policyName': f'{policy_name}.{self.__policies_domain}'
            }
        }

        if match_resources is not None:
            binding['spec']['matchResources'] = match_resources
        
        if validation_actions is not None:
            binding['spec']['validationActions'] = validation_actions

        if param_ref is not None:
            binding['spec']['paramRef'] = param_ref

        return binding
    
    def __append_bindings(self, policy_name, bindings):
        for binding in bindings:
          for key in binding:
            param_ref          = binding[key]['paramRef'] if 'paramRef' in binding[key] else None
            validation_actions = binding[key]['validationActions'] if 'validationActions' in binding[key] else None
            match_resources    = binding[key]['matchResources'] if 'matchResources' in binding[key] else None
            binding_object     = self.__create_binding(policy_name, key, param_ref, validation_actions, match_resources)

          with open(os.path.join(self.__release_output_dir, self.__bindings_file), 'a') as file:
              yaml.dump(binding_object, file, default_flow_style=False)
              file.write('---\n')
    
    def __append_crds(self, subdirectory):
        crd_path = os.path.join('../policies', subdirectory, 'crd-parameter.yaml')
        if os.path.exists(crd_path):
            with open(crd_path, 'r') as file:
                policy_content = file.read().rstrip('\n').lstrip('---\n')
  
            with open(os.path.join(self.__release_output_dir, self.__crds_file), 'a') as file:
                file.write(policy_content + "\n")
                file.write('---\n')
        else:
          print(f"Warning: {crd_path} does not exist.")

    def __append_kustomization(self, subdirectory):
        kustomization_config = {
            'apiVersion': 'kustomize.config.k8s.io/v1beta1',
            'kind': 'Kustomization',
            'resources': [self.__bindings_file, self.__crds_file, self.__policies_file]
        }

        with open(os.path.join(self.__release_output_dir, self.__kustomization_file), 'a') as file:
              yaml.dump(kustomization_config, file, default_flow_style=False)


    def generate_release_files(self):
        # Ensure the output files are empty at the start
        with open(os.path.join(self.__release_output_dir, self.__policies_file), 'w') as file:
            pass
    
        with open(os.path.join(self.__release_output_dir, self.__bindings_file), 'w') as file:
            pass
    
        with open(os.path.join(self.__release_output_dir, self.__crds_file), 'w') as file:
            pass
        
        with open(os.path.join(self.__release_output_dir, self.__kustomization_file), 'w') as file:
            pass
    
        for subdirectory, details in self.__config.items():
            # If enabled:true is present we process the entry, else we skip it
            if details.get('enabled', False) == True:
                # Add the requested policy
                self.__append_policy(subdirectory)
                bindings = details.get('bindings', {})
                # Generate the requested bindings
                if bindings:
                    self.__append_bindings(subdirectory, bindings)
                # Add CRDs for the policy if they exist
                self.__append_crds(subdirectory)
        
        # Generate kustomization config
        self.__append_kustomization(subdirectory)


def main(config_path):
    policies_file = 'policies.yaml'
    bindings_file = 'bindings.yaml'
    crds_file = 'crds.yaml'
    kustomization_file = 'kustomization.yaml'
    policies_domain = 'vap-library.com'

    new_release = Release(config_path, policies_domain, policies_file, bindings_file, crds_file, kustomization_file)
    new_release.generate_release_files()
        
    print(f"Policies have been appended to {policies_file}")
    print(f"Bindings have been appended to {bindings_file}")
    print(f"CRDs have been appended to {crds_file}, if they exist")
    print(f"Kustomization generated at {kustomization_file}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python script.py <config.yaml>")
    else:
        config_path = sys.argv[1]
        main(config_path)
